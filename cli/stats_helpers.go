package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"

	"github.com/docker/go-units"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// StatsEntry represents the statistics data collected from a container
type StatsEntry struct {
	container        string
	name             string
	id               string
	cpuPercentage    float64
	memory           float64
	memoryLimit      float64
	memoryPercentage float64
	networkRx        float64
	networkTx        float64
	blockRead        float64
	blockWrite       float64
	pidsCurrent      uint64
	err              error
}

// StatsEntryWithLock represents an entity to store containers statistics synchronously
type StatsEntryWithLock struct {
	mutex sync.Mutex
	StatsEntry
}

// Name return the name of container
func (s StatsEntry) Name() string {
	return s.name
}

// ID return the id of container
func (s StatsEntry) ID() string {
	var shortLen = 12
	if len(s.id) > shortLen {
		return s.id[:shortLen]
	}
	return s.id
}

// CPUPerc return the cpu usage percentage
func (s StatsEntry) CPUPerc() string {
	if s.err != nil {
		return fmt.Sprintf("--")
	}
	return fmt.Sprintf("%.2f%%", s.cpuPercentage)
}

// MemUsage return memory usage
func (s StatsEntry) MemUsage() string {
	if s.err != nil {
		return fmt.Sprintf("-- / --")
	}

	return fmt.Sprintf("%s / %s", units.BytesSize(s.memory), units.BytesSize(s.memoryLimit))
}

// MemPerc return memory percentage
func (s StatsEntry) MemPerc() string {
	if s.err != nil {
		return fmt.Sprintf("--")
	}
	return fmt.Sprintf("%.2f%%", s.memoryPercentage)
}

// NetIO return net IO usage of container
func (s StatsEntry) NetIO() string {
	if s.err != nil {
		return fmt.Sprintf("--")
	}
	return fmt.Sprintf("%s / %s", units.HumanSizeWithPrecision(s.networkRx, 3), units.HumanSizeWithPrecision(s.networkTx, 3))
}

// BlockIO return block IO usage of container
func (s StatsEntry) BlockIO() string {
	if s.err != nil {
		return fmt.Sprintf("--")
	}
	return fmt.Sprintf("%s / %s", units.HumanSizeWithPrecision(s.blockRead, 3), units.HumanSizeWithPrecision(s.blockWrite, 3))
}

// PIDs return current pid of container
func (s StatsEntry) PIDs() string {
	if s.err != nil {
		return fmt.Sprintf("--")
	}
	return fmt.Sprintf("%d", s.pidsCurrent)
}

// GetStatsEntry return the StatsEntry of StatsEntryWithLock
func (s *StatsEntryWithLock) GetStatsEntry() StatsEntry {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.StatsEntry
}

// SetError will set error of StatsEntryWithLock
func (s *StatsEntryWithLock) SetError(err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.err = err
}

// GetError will return the error string of StatsEntryWithLock
func (s *StatsEntryWithLock) GetError() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.err != nil {
		return fmt.Errorf("failed to stats container %s, err = %v", s.id, s.err.Error())
	}
	return nil
}

func collect(ctx context.Context, s *StatsEntryWithLock, cli client.CommonAPIClient, streamStats bool, waitFirst *sync.WaitGroup) {
	logrus.Debugf("collecting stats for %s", s.container)
	var (
		getFirst       bool
		previousCPU    uint64
		previousSystem uint64
		errCh          = make(chan error, 1)
		dataCh         = make(chan struct{}, 1)
	)

	defer func() {
		// double check
		if !getFirst {
			getFirst = true
			waitFirst.Done()
		}
	}()

	response, err := cli.ContainerStats(ctx, s.container, streamStats)
	if err != nil {
		s.SetError(err)
		return
	}
	defer response.Close()

	dec := json.NewDecoder(response)
	go func() {
		for {
			var (
				v                      *types.ContainerStats
				memPercent, cpuPercent float64
				blkRead, blkWrite      uint64
				mem, memLimit          float64
				pidsStatsCurrent       uint64
			)

			if err := dec.Decode(&v); err != nil {
				dec = json.NewDecoder(io.MultiReader(dec.Buffered(), response))
				errCh <- err
				break
			}

			// first time PrecpuStats will be nil
			if v.PrecpuStats != nil {
				previousCPU = v.PrecpuStats.CPUUsage.TotalUsage
				previousSystem = v.PrecpuStats.SyetemCPUUsage
			}
			cpuPercent = calculateCPUPercentUnix(previousCPU, previousSystem, v)
			blkRead, blkWrite = calculateBlockIO(v.BlkioStats)
			mem = calculateMemUsageUnixNoCache(v.MemoryStats)
			memLimit = float64(v.MemoryStats.Limit)
			memPercent = calculateMemPercentUnixNoCache(memLimit, mem)
			pidsStatsCurrent = v.PidsStats.Current
			netRx, netTx := calculateNetwork(v.Networks)

			s.mutex.Lock()
			s.name = v.Name
			s.id = v.ID
			s.cpuPercentage = cpuPercent
			s.memory = mem
			s.memoryLimit = memLimit
			s.memoryPercentage = memPercent
			s.networkRx = netRx
			s.networkTx = netTx
			s.blockRead = float64(blkRead)
			s.blockWrite = float64(blkWrite)
			s.pidsCurrent = pidsStatsCurrent
			s.mutex.Unlock()

			dataCh <- struct{}{}

			if !streamStats {
				return
			}
		}
	}()

	for {
		select {
		case <-time.After(2 * time.Second):
			// zero out the values if we have not received an update within
			// the specified duration.
			s.SetError(errors.New("timeout waiting for stats"))
			//FIXME(ZYecho): should retry when timeout?
			return
		case err := <-errCh:
			s.SetError(err)
			return
		case <-dataCh:
			// if this is the first stat you get, release WaitGroup
			if !getFirst {
				getFirst = true
				waitFirst.Done()
			}
			if !streamStats {
				return
			}
		}
	}
}

func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v *types.ContainerStats) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SyetemCPUUsage) - float64(previousSystem)
		onlineCPUs  = float64(v.CPUStats.OnlineCpus)
	)

	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(v.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}
	return cpuPercent
}

func calculateBlockIO(blkio *types.BlkioStats) (uint64, uint64) {
	var blkRead, blkWrite uint64
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = blkRead + bioEntry.Value
		case "write":
			blkWrite = blkWrite + bioEntry.Value
		}
	}
	return blkRead, blkWrite
}

func calculateNetwork(network map[string]types.NetworkStats) (float64, float64) {
	var rx, tx float64

	for _, v := range network {
		rx += float64(v.RxBytes)
		tx += float64(v.TxBytes)
	}
	return rx, tx
}

// calculateMemUsageUnixNoCache calculate memory usage of the container.
// Page cache is intentionally excluded to avoid misinterpretation of the output.
func calculateMemUsageUnixNoCache(mem *types.MemoryStats) float64 {
	return float64(mem.Usage - mem.Stats["cache"])
}

func calculateMemPercentUnixNoCache(limit float64, usedNoCache float64) float64 {
	// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
	// got any data from cgroup
	if limit != 0 {
		return usedNoCache / limit * 100.0
	}
	return 0
}
