package mgr

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alibaba/pouch/apis/types"

	"github.com/containerd/cgroups"
	containerdtypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/typeurl"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/go-openapi/strfmt"
	"github.com/sirupsen/logrus"
)

// StreamStats gets the stats from containerd side and send back to caller as a stream.
func (mgr *ContainerManager) StreamStats(ctx context.Context, name string, config *ContainerStatsConfig) error {
	c, err := mgr.container(name)
	if err != nil {
		return err
	}

	outStream := config.OutStream
	if (!c.IsRunning() || c.IsRestarting()) && !config.Stream {
		return json.NewEncoder(outStream).Encode(&types.ContainerStats{
			Name: c.Name,
			ID:   c.ID,
		})
	}

	if c.IsRunning() && !config.Stream {
		metrics, stats, err := mgr.Stats(ctx, name)
		if err != nil {
			return err
		}
		containerStat := toContainerStats(metrics.Timestamp, stats)
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
			containerStat := toContainerStats(metrics.Timestamp, stats)

			if err := enc.Encode(containerStat); err != nil {
				return err
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
	ID := c.ID
	c.Unlock()

	metric, err := mgr.Client.ContainerStats(ctx, ID)
	if err != nil {
		return nil, nil, err
	}

	v, err := typeurl.UnmarshalAny(metric.Data)
	if err != nil {
		return nil, nil, err
	}

	return metric, v.(*cgroups.Metrics), nil
}

func toContainerStats(time time.Time, metric *cgroups.Metrics) *types.ContainerStats {
	return &types.ContainerStats{
		Read: strfmt.DateTime(time),
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
			// Add SyetemCPUUsage?
			// SyetemCPUUsage: metric.CPU.Usage.SyetemCPUUsage,
			ThrottlingData: &types.ThrottlingData{
				Periods:          metric.CPU.Throttling.Periods,
				ThrottledPeriods: metric.CPU.Throttling.ThrottledPeriods,
				ThrottledTime:    metric.CPU.Throttling.ThrottledTime,
			},
		},
		PrecpuStats: &types.CPUStats{},
		BlkioStats:  &types.BlkioStats{},
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
