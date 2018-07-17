package mgr

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/logger/syslog"
	"github.com/alibaba/pouch/pkg/system"

	"github.com/sirupsen/logrus"
)

// validateConfig validates container config
func (mgr *ContainerManager) validateConfig(c *Container, update bool) ([]string, error) {
	// validates container hostconfig
	hostConfig := c.HostConfig
	warnings := make([]string, 0)
	warns, err := validateResource(&hostConfig.Resources, update)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, warns...)

	if hostConfig.OomScoreAdj < -1000 || hostConfig.OomScoreAdj > 1000 {
		return warnings, fmt.Errorf("oom score should be in range [-1000, 1000]")
	}

	if hostConfig.ShmSize != nil && *hostConfig.ShmSize < 0 {
		return warnings, fmt.Errorf("shm-size %d should greater than 0", *hostConfig.ShmSize)
	}

	// validate log config
	if err := mgr.validateLogConfig(c); err != nil {
		return warnings, err
	}

	// TODO: validate config
	return warnings, nil
}

// validateResource verifies cgroup resources
func validateResource(r *types.Resources, update bool) ([]string, error) {
	cgroupInfo := system.NewCgroupInfo()
	if cgroupInfo == nil {
		return nil, nil
	}
	warnings := make([]string, 0, 64)

	// validates memory cgroup value
	if cgroupInfo.Memory != nil {
		if r.Memory > 0 && !cgroupInfo.Memory.MemoryLimit {
			logrus.Warn(MemoryWarn)
			warnings = append(warnings, MemoryWarn)
			r.Memory = 0
			r.MemorySwap = 0
		}
		if r.MemorySwap > 0 && !cgroupInfo.Memory.MemorySwap {
			logrus.Warn(MemorySwapWarn)
			warnings = append(warnings, MemorySwapWarn)
			r.MemorySwap = 0
		}
		// cgroup not allow memory-swap less than memory limit
		if r.Memory > 0 && r.MemorySwap > 0 && r.MemorySwap < r.Memory {
			return warnings, fmt.Errorf("Minimum memoryswap limit should be larger than memory limit")
		}
		// cgroup not allow set memory-swap without set memory
		if r.Memory == 0 && r.MemorySwap > 0 && !update {
			return warnings, fmt.Errorf("You should always set the Memory limit when using Memoryswap limit")
		}
		if r.Memory != 0 && r.Memory < MinMemory {
			return warnings, fmt.Errorf("Minimal memory should greater than 4M")
		}
		if r.Memory > 0 && r.MemorySwap > 0 && r.MemorySwap < 2*r.Memory {
			warnings = append(warnings, "You should typically size your swap space to approximately 2x main memory for systems with less than 2GB of RAM")
		}
		if r.MemorySwappiness != nil && !cgroupInfo.Memory.MemorySwappiness {
			logrus.Warn(MemorySwappinessWarn)
			warnings = append(warnings, MemorySwappinessWarn)
			r.MemorySwappiness = nil
		}
		if r.MemorySwappiness != nil && (*r.MemorySwappiness < 0 || *r.MemorySwappiness > 100) {
			return warnings, fmt.Errorf("MemorySwappiness should in range [0, 100]")
		}
		if r.OomKillDisable != nil && !cgroupInfo.Memory.OOMKillDisable {
			logrus.Warn(OOMKillWarn)
			warnings = append(warnings, OOMKillWarn)
			r.OomKillDisable = nil
		}
	}

	// validates cpu cgroup value
	if cgroupInfo.CPU != nil {
		if r.CpusetCpus != "" && !cgroupInfo.CPU.CpusetCpus {
			logrus.Warn(CpusetCpusWarn)
			warnings = append(warnings, CpusetCpusWarn)
			r.CpusetCpus = ""
		}
		if r.CpusetMems != "" && !cgroupInfo.CPU.CpusetMems {
			logrus.Warn(CpusetMemsWarn)
			warnings = append(warnings, CpusetMemsWarn)
			r.CpusetMems = ""
		}
		if r.CPUShares > 0 && !cgroupInfo.CPU.CPUShares {
			logrus.Warn(CPUSharesWarn)
			warnings = append(warnings, CPUSharesWarn)
			r.CPUShares = 0
		}
		if r.CPUQuota > 0 && !cgroupInfo.CPU.CPUQuota {
			logrus.Warn(CPUQuotaWarn)
			warnings = append(warnings, CPUQuotaWarn)
			r.CPUQuota = 0
		}
		// cpu.cfs_quota_us can accept value less than 0, we allow -1 and > 1000
		if r.CPUQuota > 0 && r.CPUQuota < 1000 {
			return warnings, fmt.Errorf("CPU cfs quota should be greater than 1ms(1000)")
		}
		if r.CPUPeriod > 0 && !cgroupInfo.CPU.CPUPeriod {
			logrus.Warn(CPUPeriodWarn)
			warnings = append(warnings, CPUPeriodWarn)
			r.CPUPeriod = 0
		}
		if r.CPUPeriod != 0 && (r.CPUPeriod < 1000 || r.CPUPeriod > 1000000) {
			return warnings, fmt.Errorf("CPU cfs period should be in range [1000, 1000000](1ms, 1s)")
		}
	}

	// validates blkio cgroup value
	if cgroupInfo.Blkio != nil {
		if r.BlkioWeight > 0 && !cgroupInfo.Blkio.BlkioWeight {
			logrus.Warn(BlkioWeightWarn)
			warnings = append(warnings, BlkioWeightWarn)
			r.BlkioWeight = 0
		}
		if len(r.BlkioWeightDevice) > 0 && !cgroupInfo.Blkio.BlkioWeightDevice {
			logrus.Warn(BlkioWeightDeviceWarn)
			warnings = append(warnings, BlkioWeightDeviceWarn)
			r.BlkioWeightDevice = []*types.WeightDevice{}
		}
		if len(r.BlkioDeviceReadBps) > 0 && !cgroupInfo.Blkio.BlkioDeviceReadBps {
			logrus.Warn(BlkioDeviceReadBpsWarn)
			warnings = append(warnings, BlkioDeviceReadBpsWarn)
			r.BlkioDeviceReadBps = []*types.ThrottleDevice{}
		}
		if len(r.BlkioDeviceWriteBps) > 0 && !cgroupInfo.Blkio.BlkioDeviceWriteBps {
			logrus.Warn(BlkioDeviceWriteBpsWarn)
			warnings = append(warnings, BlkioDeviceWriteBpsWarn)
			r.BlkioDeviceWriteBps = []*types.ThrottleDevice{}
		}
		if len(r.BlkioDeviceReadIOps) > 0 && !cgroupInfo.Blkio.BlkioDeviceReadIOps {
			logrus.Warn(BlkioDeviceReadIOpsWarn)
			warnings = append(warnings, BlkioDeviceReadIOpsWarn)
			r.BlkioDeviceReadIOps = []*types.ThrottleDevice{}
		}
		if len(r.BlkioDeviceWriteIOps) > 0 && !cgroupInfo.Blkio.BlkioDeviceWriteIOps {
			logrus.Warn(BlkioDeviceWriteIOpsWarn)
			warnings = append(warnings, BlkioDeviceWriteIOpsWarn)
			r.BlkioDeviceWriteIOps = []*types.ThrottleDevice{}
		}
	}

	// validates pid cgroup value
	if cgroupInfo.Pids != nil {
		if r.PidsLimit != 0 && !cgroupInfo.Pids.Pids {
			logrus.Warn(PidsLimitWarn)
			warnings = append(warnings, PidsLimitWarn)
			r.PidsLimit = 0
		}
	}

	return warnings, nil
}

// validateLogConfig is used to verify the correctness of log configuration.
// TODO(fuwei): remove mgr from validateLogConfig
func (mgr *ContainerManager) validateLogConfig(c *Container) error {
	logCfg := c.HostConfig.LogConfig
	if logCfg == nil {
		return nil
	}

	switch logCfg.LogDriver {
	case types.LogConfigLogDriverNone, types.LogConfigLogDriverJSONFile:
		return nil
	case types.LogConfigLogDriverSyslog:
		info := mgr.convContainerToLoggerInfo(c)
		return syslog.ValidateSyslogOption(info)
	default:
		return fmt.Errorf("not support (%v) log driver yet", logCfg.LogDriver)
	}
}
