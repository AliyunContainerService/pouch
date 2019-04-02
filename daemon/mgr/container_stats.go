package mgr

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"

	"github.com/containerd/cgroups"
	containerdtypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/typeurl"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/go-openapi/strfmt"
	"github.com/opencontainers/runc/libcontainer/system"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const nanoSecondsPerSecond = 1e9

// StreamStats gets the stats from containerd side and send back to caller as a stream.
func (mgr *ContainerManager) StreamStats(ctx context.Context, name string, config *ContainerStatsConfig) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	outStream := config.OutStream
	if !c.IsRunningOrPaused() {
		return errors.New("can only stats running or paused container")
	}

	var preCPUStats *types.CPUStats

	wrapContainerStats := func(timestamp time.Time, metric *cgroups.Metrics) (*types.ContainerStats, error) {
		stats := toContainerStats(timestamp, c, metric)
		systemCPUUsage, err := getSystemCPUUsage()
		if err != nil {
			return nil, err
		}
		stats.PrecpuStats = preCPUStats
		stats.CPUStats.SyetemCPUUsage = systemCPUUsage
		preCPUStats = stats.CPUStats
		networkStat, err := mgr.NetworkMgr.GetNetworkStats(c.NetworkSettings.SandboxID)
		if err != nil {
			// --net=none or disconnect from network, the sandbox will be nil
			logrus.Warnf("sandbox not found, name = %s, err= %v", name, err)
		}
		stats.Networks = networkStat
		return stats, nil
	}

	if c.IsRunningOrPaused() && !config.Stream {
		metrics, stats, err := mgr.Stats(ctx, name)
		if err != nil {
			return err
		}
		containerStat, err := wrapContainerStats(metrics.Timestamp, stats)
		if err != nil {
			return errors.Errorf("failed to wrap the containerStat: %v", err)
		}
		return json.NewEncoder(outStream).Encode(containerStat)
	}

	if config.Stream {
		wf := ioutils.NewWriteFlusher(outStream)
		defer wf.Close()
		wf.Flush()
		outStream = wf
	}

	enc := json.NewEncoder(outStream)

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("context is cancelled when streaming stats of container %s", c.ID)
			return nil
		default:
			logrus.Debugf("Start to stream stats of container %s", c.ID)
			metrics, stats, err := mgr.Stats(ctx, name)
			if err != nil {
				return err
			}

			if metrics != nil {
				containerStat, err := wrapContainerStats(metrics.Timestamp, stats)
				if err != nil {
					return errors.Errorf("failed to wrap the containerStat: %v", err)
				}
				if err := enc.Encode(containerStat); err != nil {
					return err
				}
			}

			time.Sleep(DefaultStatsInterval)
		}
	}
}

// Stats gets the stat of a container.
func (mgr *ContainerManager) Stats(ctx context.Context, name string) (*containerdtypes.Metric, *cgroups.Metrics, error) {
	c, err := mgr.container(name)
	if err != nil {
		return nil, nil, err
	}

	c.Lock()
	defer c.Unlock()

	// only get metrics when the container is running
	// return error to help client quick fail
	if !c.IsRunningOrPaused() {
		return nil, nil, errors.New("can only stats running or paused container")
	}

	metric, err := mgr.Client.ContainerStats(ctx, c.ID)
	if err != nil {
		return nil, nil, err
	}

	v, err := typeurl.UnmarshalAny(metric.Data)
	if err != nil {
		return nil, nil, err
	}

	return metric, v.(*cgroups.Metrics), nil
}

func toContainerStats(time time.Time, container *Container, metric *cgroups.Metrics) *types.ContainerStats {
	return &types.ContainerStats{
		Read: strfmt.DateTime(time),
		ID:   container.ID,
		Name: container.Name,
		PidsStats: &types.PidsStats{
			Current: metric.Pids.Current,
		},
		CPUStats: &types.CPUStats{
			CPUUsage: &types.CPUUsage{
				PercpuUsage:       metric.CPU.Usage.PerCPU,
				TotalUsage:        metric.CPU.Usage.Total,
				UsageInKernelmode: metric.CPU.Usage.Kernel,
				UsageInUsermode:   metric.CPU.Usage.User,
			},
			ThrottlingData: &types.ThrottlingData{
				Periods:          metric.CPU.Throttling.Periods,
				ThrottledPeriods: metric.CPU.Throttling.ThrottledPeriods,
				ThrottledTime:    metric.CPU.Throttling.ThrottledTime,
			},
		},
		BlkioStats: &types.BlkioStats{
			IoServiceBytesRecursive: toContainerBlkioStatsEntry(metric.Blkio.IoServiceBytesRecursive),
			IoServicedRecursive:     toContainerBlkioStatsEntry(metric.Blkio.IoServicedRecursive),
			IoQueueRecursive:        toContainerBlkioStatsEntry(metric.Blkio.IoQueuedRecursive),
			IoServiceTimeRecursive:  toContainerBlkioStatsEntry(metric.Blkio.IoServiceTimeRecursive),
			IoWaitTimeRecursive:     toContainerBlkioStatsEntry(metric.Blkio.IoWaitTimeRecursive),
			IoMergedRecursive:       toContainerBlkioStatsEntry(metric.Blkio.IoMergedRecursive),
			IoTimeRecursive:         toContainerBlkioStatsEntry(metric.Blkio.IoTimeRecursive),
			SectorsRecursive:        toContainerBlkioStatsEntry(metric.Blkio.SectorsRecursive),
		},
		MemoryStats: &types.MemoryStats{
			Stats: map[string]uint64{
				"total_pgmajfault":          metric.Memory.TotalPgMajFault,
				"cache":                     metric.Memory.Cache,
				"mapped_file":               metric.Memory.MappedFile,
				"total_inactive_file":       metric.Memory.TotalInactiveFile,
				"pgpgout":                   metric.Memory.PgPgOut,
				"rss":                       metric.Memory.RSS,
				"total_mapped_file":         metric.Memory.TotalMappedFile,
				"writeback":                 metric.Memory.Writeback,
				"unevictable":               metric.Memory.Unevictable,
				"pgpgin":                    metric.Memory.PgPgIn,
				"total_unevictable":         metric.Memory.TotalUnevictable,
				"pgmajfault":                metric.Memory.PgMajFault,
				"total_rss":                 metric.Memory.TotalRSS,
				"total_rss_huge":            metric.Memory.TotalRSSHuge,
				"total_writeback":           metric.Memory.TotalWriteback,
				"total_inactive_anon":       metric.Memory.TotalInactiveAnon,
				"rss_huge":                  metric.Memory.RSSHuge,
				"hierarchical_memory_limit": metric.Memory.HierarchicalMemoryLimit,
				"total_pgfault":             metric.Memory.TotalPgFault,
				"total_active_file":         metric.Memory.TotalActiveFile,
				"active_anon":               metric.Memory.ActiveAnon,
				"total_active_anon":         metric.Memory.TotalActiveAnon,
				"total_pgpgout":             metric.Memory.TotalPgPgOut,
				"total_cache":               metric.Memory.TotalCache,
				"inactive_anon":             metric.Memory.InactiveAnon,
				"active_file":               metric.Memory.ActiveFile,
				"pgfault":                   metric.Memory.PgFault,
				"inactive_file":             metric.Memory.InactiveFile,
				"total_pgpgin":              metric.Memory.PgPgIn,
			},
			Usage:    metric.Memory.Usage.Usage,
			Failcnt:  metric.Memory.Usage.Failcnt,
			MaxUsage: metric.Memory.Usage.Max,
			Limit:    metric.Memory.Usage.Limit,
		},
	}
}

func toContainerBlkioStatsEntry(statEntrys []*cgroups.BlkIOEntry) []*types.BlkioStatEntry {
	blkioStatEntrys := []*types.BlkioStatEntry{}
	for _, item := range statEntrys {
		blkioStatEntrys = append(blkioStatEntrys, &types.BlkioStatEntry{
			Major: item.Major,
			Minor: item.Minor,
			Op:    item.Op,
			Value: item.Value,
		})
	}
	return blkioStatEntrys
}

// getSystemCPUUsage returns the host system's cpu usage in
// nanoseconds. An error is returned if the format of the underlying
// file does not match.
//
// Uses /proc/stat defined by POSIX. Looks for the cpu
// statistics line and then sums up the first seven fields
// provided. See `man 5 proc` for details on specific field
// information.
func getSystemCPUUsage() (uint64, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	bufReader := bufio.NewReaderSize(nil, 128)
	defer func() {
		bufReader.Reset(nil)
		f.Close()
	}()
	bufReader.Reset(f)

	for {
		line, err := bufReader.ReadString('\n')
		if err != nil {
			break
		}
		parts := strings.Fields(line)
		switch parts[0] {
		case "cpu":
			if len(parts) < 8 {
				return 0, fmt.Errorf("invalid number of cpu fields")
			}
			var totalClockTicks uint64
			for _, i := range parts[1:8] {
				v, err := strconv.ParseUint(i, 10, 64)
				if err != nil {
					return 0, fmt.Errorf("unable to convert value %s to int: %s", i, err)
				}
				totalClockTicks += v
			}
			return (totalClockTicks * nanoSecondsPerSecond) /
				uint64(system.GetClockTicks()), nil
		}
	}
	return 0, fmt.Errorf("invalid stat format, fail to parse the '/proc/stat' file")
}
