package system

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// MemoryCgroupInfo defines memory cgroup information on current machine
type MemoryCgroupInfo struct {
	MemoryLimit      bool
	MemorySwap       bool
	MemorySwappiness bool
	OOMKillDisable   bool
}

// CPUCgroupInfo defines cpu cgroup information on current machine
type CPUCgroupInfo struct {
	CpusetCpus bool
	CpusetMems bool
	CPUShares  bool
	CPUPeriod  bool
	CPUQuota   bool
}

// BlkioCgroupInfo defines blkio cgroup information on current machine
type BlkioCgroupInfo struct {
	BlkioWeight          bool
	BlkioWeightDevice    bool
	BlkioDeviceReadBps   bool
	BlkioDeviceWriteBps  bool
	BlkioDeviceReadIOps  bool
	BlkioDeviceWriteIOps bool
}

// PidsCgroupInfo defines pid cgroup information on current machine
type PidsCgroupInfo struct {
	Pids bool
}

// CgroupInfo defines cgroup information on current machine
type CgroupInfo struct {
	Memory *MemoryCgroupInfo
	CPU    *CPUCgroupInfo
	Blkio  *BlkioCgroupInfo
	Pids   *PidsCgroupInfo
}

// NewCgroupInfo news a CgroupInfo struct
func NewCgroupInfo() *CgroupInfo {
	cgroupRootPath := getCgroupRootMount("/proc/self/mountinfo")
	if cgroupRootPath == "" {
		return nil
	}

	return &CgroupInfo{
		Memory: getMemoryCgroupInfo(cgroupRootPath),
		CPU:    getCPUCgroupInfo(cgroupRootPath),
		Blkio:  getBlkioCgroupInfo(cgroupRootPath),
		Pids:   getPidsCgroupInfo(cgroupRootPath),
	}
}

func getMemoryCgroupInfo(root string) *MemoryCgroupInfo {
	path := path.Join(root, "memory")
	return &MemoryCgroupInfo{
		MemoryLimit:      isCgroupEnable(path, "memory.limit_in_bytes"),
		MemorySwap:       isCgroupEnable(path, "memory.memsw.limit_in_bytes"),
		MemorySwappiness: isCgroupEnable(path, "memory.swappiness"),
		OOMKillDisable:   isCgroupEnable(path, "memory.oom_control"),
	}
}

func getCPUCgroupInfo(root string) *CPUCgroupInfo {
	cpuPath := path.Join(root, "cpu")
	cpusetPath := path.Join(root, "cpuset")
	return &CPUCgroupInfo{
		CpusetCpus: isCgroupEnable(cpusetPath, "cpuset.cpus"),
		CpusetMems: isCgroupEnable(cpusetPath, "cpuset.mems"),
		CPUShares:  isCgroupEnable(cpuPath, "cpu.shares"),
		CPUQuota:   isCgroupEnable(cpuPath, "cpu.cfs_quota_us"),
		CPUPeriod:  isCgroupEnable(cpuPath, "cpu.cfs_period_us"),
	}
}

func getBlkioCgroupInfo(root string) *BlkioCgroupInfo {
	path := path.Join(root, "blkio")
	return &BlkioCgroupInfo{
		BlkioWeight:          isCgroupEnable(path, "blkio.weight"),
		BlkioWeightDevice:    isCgroupEnable(path, "blkio.weight_device"),
		BlkioDeviceReadBps:   isCgroupEnable(path, "blkio.throttle.read_bps_device"),
		BlkioDeviceWriteBps:  isCgroupEnable(path, "blkio.throttle.write_bps_device"),
		BlkioDeviceReadIOps:  isCgroupEnable(path, "blkio.throttle.read_iops_device"),
		BlkioDeviceWriteIOps: isCgroupEnable(path, "blkio.throttle.write_iops_device"),
	}
}

func getPidsCgroupInfo(root string) *PidsCgroupInfo {
	return &PidsCgroupInfo{
		Pids: isCgroupEnable(path.Join(root, "pids")),
	}
}

func isCgroupEnable(f ...string) bool {
	_, exist := os.Stat(path.Join(f...))
	return exist == nil
}

func getCgroupRootMount(mountFile string) string {
	f, err := os.Open(mountFile)
	if err != nil {
		return ""
	}
	defer f.Close()

	var cgroupRootPath string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		index := strings.Index(text, " - ")
		if index < 0 {
			continue
		}
		fields := strings.Split(text, " ")
		postSeparatorFields := strings.Fields(text[index+3:])
		numPostFields := len(postSeparatorFields)

		if len(fields) < 5 || postSeparatorFields[0] != "cgroup" || numPostFields < 3 {
			continue
		}

		cgroupRootPath = filepath.Dir(fields[4])
		break
	}

	if _, err = os.Stat(cgroupRootPath); err != nil {
		return ""
	}

	return cgroupRootPath
}
