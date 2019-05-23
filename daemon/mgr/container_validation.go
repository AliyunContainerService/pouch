package mgr

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/daemon/logger/jsonfile"
	"github.com/alibaba/pouch/daemon/logger/syslog"
	"github.com/alibaba/pouch/pkg/system"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/storage/quota"

	"github.com/docker/go-units"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// all: all GPUs will be accessible, this is the default value in our container images.
	// none: no GPU will be accessible, but driver capabilities will be enabled.
	supportedDevices = map[string]*struct{}{"all": nil, "none": nil, "void": nil}

	// none: no GPU will be accessible, but driver capabilities will be enabled.
	// void or empty: no GPU will be accessible, and driver capabilities will be disabled.
	// all: all GPUs will be accessible
	supportedDrivers = map[string]*struct{}{"compute": nil, "compat32": nil, "graphics": nil, "utility": nil, "video": nil, "display": nil}

	errInvalidDevice    = errors.New("invalid nvidia device")
	errInvalidDriver    = errors.New("invalid nvidia driver capability")
	errInvalidDiskQuota = errors.New("invalid disk quota")

	// commonLogOpts the option which should be validated in common such as mode, max-buffer-size.
	commonLogOpts = map[string]bool{
		"mode":            true,
		"max-buffer-size": true,
		logRootDirKey:     true,
	}
)

// validateConfig validates container config
func (mgr *ContainerManager) validateConfig(c *Container, update bool) ([]string, error) {
	// validates rich mode
	if err := validateRichMode(c); err != nil {
		return nil, err
	}

	// validates container hostconfig
	hostConfig := c.HostConfig
	warnings := make([]string, 0)
	warns, err := validateResource(&hostConfig.Resources, update)
	if err != nil {
		return nil, err
	}
	// validates nvidia config
	if err := validateNvidiaConfig(&hostConfig.Resources); err != nil {
		return warnings, err
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

	// validate seccomp, apparmor security parameters
	sysInfo := system.NewInfo()
	if !sysInfo.Seccomp {
		if c.SeccompProfile != "" || c.SeccompProfile != ProfileUnconfined {
			warnings = append(warnings, fmt.Sprintf("Current Kernel does not support seccomp, discard --security-opt seccomp=%s", c.SeccompProfile))
		}
		// always set SeccompProfile to unconfined if kernel not support seccomp
		c.SeccompProfile = ProfileUnconfined

	}
	if !sysInfo.AppArmor {
		if c.AppArmorProfile != "" {
			warnings = append(warnings, fmt.Sprintf("Current Kernel does not support apparmor, discard --security-opt apparmor=%s", c.AppArmorProfile))
		}
		c.AppArmorProfile = ""
	}

	return warnings, nil
}

// validateDiskQuota is used to validate disk quota config
func (mgr *ContainerManager) validateDiskQuota(config *types.ContainerCreateConfig) error {
	if config == nil {
		return errors.Errorf("invalid request, create config is nil")
	}

	if config.DiskQuota == nil {
		if quota.IsSetQuotaID(config.QuotaID) {
			return errors.Wrap(errInvalidDiskQuota, "set QuotaID without DiskQuota")
		}
		return nil
	}

	quotaMaps := config.DiskQuota
	if len(quotaMaps) > 1 && quota.IsSetQuotaID(config.QuotaID) {
		return errors.Wrap(errInvalidDiskQuota, `QuotaID only used to set one disk quota, `+
			`such as: "/=10G" or "/path1=10G" or ".*=10G"`)
	}

	for key := range quotaMaps {
		if key == "" {
			return errors.Wrap(errInvalidDiskQuota, "quota can not be nil string")
		}

		paths := strings.Split(key, "&")
		if len(paths) <= 1 {
			continue
		}

		for _, path := range paths {
			if !filepath.IsAbs(path) {
				return errors.Wrapf(errInvalidDiskQuota,
					"(%s) is invalid path in set quota(%s)", path, key)
			}
		}
	}

	return nil
}

// validateRichMode verifies rich mode parameters
func validateRichMode(c *Container) error {
	richModes := []string{
		types.ContainerConfigRichModeDumbInit,
		types.ContainerConfigRichModeSbinInit,
		types.ContainerConfigRichModeSystemd,
	}

	if !c.Config.Rich && c.Config.RichMode != "" {
		return fmt.Errorf("must first enable rich mode, then specify a rich mode type")
	}
	// check rich mode
	if c.Config.RichMode != "" && !utils.StringInSlice(richModes, c.Config.RichMode) {
		return fmt.Errorf("not supported rich mode: %v", c.Config.RichMode)
	}

	// must use privileged when use systemd rich mode
	if c.Config.RichMode == types.ContainerConfigRichModeSystemd && !c.HostConfig.Privileged {
		return fmt.Errorf("must using privileged mode when create systemd rich container")
	}

	// if enables rich mode but not specified rich mode type, assign to dumb-init
	if c.Config.Rich && c.Config.RichMode == "" {
		c.Config.RichMode = types.ContainerConfigRichModeDumbInit
	}

	return nil
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
		if r.MemoryReservation > 0 && !cgroupInfo.Memory.MemoryReservation {
			logrus.Warn(MemoryReservationWarn)
			warnings = append(warnings, MemoryReservationWarn)
			r.MemoryReservation = 0
		}
		if r.MemoryReservation != 0 && r.MemoryReservation < MinMemory {
			return warnings, fmt.Errorf("Minimal memory reservation should greater than 4M")
		}
		if r.Memory > 0 && r.MemoryReservation > 0 && r.Memory < r.MemoryReservation {
			return warnings, fmt.Errorf("Minimum memory limit should be larger than memory reservation limit")
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
		if r.MemorySwappiness != nil && *r.MemorySwappiness != -1 && (*r.MemorySwappiness < 0 || *r.MemorySwappiness > 100) {
			return warnings, fmt.Errorf("MemorySwappiness should in range [0, 100] or -1 as a legacy alias of 0")
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

	// validate log mode
	switch logger.LogMode(logCfg.LogOpts["mode"]) {
	case logger.LogModeDefault, logger.LogModeBlocking, logger.LogModeNonBlock:
	default:
		return fmt.Errorf("unsupported logger mode: %s", logCfg.LogOpts["mode"])
	}

	// validate max buffer size of logger
	if maxBufferSize, ok := logCfg.LogOpts["max-buffer-size"]; ok {
		if logger.LogMode(logCfg.LogOpts["mode"]) != logger.LogModeNonBlock {
			return fmt.Errorf("max-buffer-size option is only supported with 'mode=%s'", logger.LogModeNonBlock)
		}

		// try to parse the max-buffer-size option
		_, err := units.RAMInBytes(maxBufferSize)
		if err != nil {
			return errors.Wrapf(err, "failed to parse option max-buffer-size: %s", maxBufferSize)
		}
	}

	// filter the option which have been validated in common.
	restOpts := make(map[string]string)
	for k, v := range logCfg.LogOpts {
		if !commonLogOpts[k] {
			restOpts[k] = v
		}
	}

	switch logCfg.LogDriver {
	case types.LogConfigLogDriverNone, types.LogConfigLogDriverJSONFile:
		return jsonfile.ValidateLogOpt(restOpts)
	case types.LogConfigLogDriverSyslog:
		info, err := mgr.convContainerToLoggerInfo(c)
		if err != nil {
			return err
		}
		return syslog.ValidateSyslogOption(info)
	default:
		return fmt.Errorf("not support (%v) log driver yet", logCfg.LogDriver)
	}
}

// validateNvidiaConfig
func validateNvidiaConfig(r *types.Resources) error {
	if r.NvidiaConfig == nil {
		return nil
	}

	if err := validateNvidiaDriver(r); err != nil {
		return err
	}

	return validateNvidiaDevice(r)
}

func validateNvidiaDriver(r *types.Resources) error {
	n := r.NvidiaConfig
	n.NvidiaDriverCapabilities = strings.TrimSpace(n.NvidiaDriverCapabilities)

	if n.NvidiaDriverCapabilities == "" {
		// use default driver capability: utility
		return nil
	}

	if n.NvidiaDriverCapabilities == "all" {
		// enable all capabilities
		return nil
	}

	drivers := strings.Split(n.NvidiaDriverCapabilities, ",")

	for _, d := range drivers {
		d = strings.TrimSpace(d)
		if _, found := supportedDrivers[d]; !found {
			return errInvalidDriver
		}
	}
	return nil
}

func validateNvidiaDevice(r *types.Resources) error {
	n := r.NvidiaConfig
	n.NvidiaVisibleDevices = strings.TrimSpace(n.NvidiaVisibleDevices)

	if n.NvidiaVisibleDevices == "" {
		// no GPU will be accessible, and driver capabilities will be disabled.
		return nil
	}

	if _, found := supportedDevices[n.NvidiaVisibleDevices]; found {
		return nil
	}

	// 0,1,2, GPU-fef8089b …: a comma-separated list of GPU UUID(s) or index(es).
	devs := strings.Split(n.NvidiaVisibleDevices, ",")
	for _, dev := range devs {
		dev = strings.TrimSpace(dev)
		if _, err := strconv.Atoi(dev); err == nil {
			//dev is numeric, the realDev should be /dev/nvidiaN
			realDev := fmt.Sprintf("/dev/nvidia%s", dev)
			if _, err := os.Stat(realDev); err != nil {
				return errInvalidDevice
			}
		}
		// TODO: how to validate GPU UUID
	}
	return nil
}
