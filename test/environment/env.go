package environment

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/system"

	"github.com/gotestyourself/gotestyourself/icmd"
)

var (
	// PouchBinary is default binary
	PouchBinary = "/usr/local/bin/pouch"

	// PouchdAddress is default pouchd address
	PouchdAddress = "unix:///var/run/pouchd.sock"

	// PouchdUnixDomainSock is the default unix domain socket file used by pouchd.
	PouchdUnixDomainSock = "/var/run/pouchd.sock"

	// TLSConfig is default tls config
	TLSConfig = client.TLSConfig{}

	// BusyboxRepo the repository of busybox image
	BusyboxRepo = "registry.hub.docker.com/library/busybox"

	// BusyboxTag the default tag used for busybox image
	BusyboxTag = "1.28"

	// BusyboxDigest the digest used for busybox image
	BusyboxDigest = "sha256:141c253bc4c3fd0a201d32dc1f493bcf3fff003b6df416dea4f41046e0f37d47"

	// BusyboxID the default ID for busybox image
	BusyboxID = "sha256:8c811b4aec35f259572d0f79207bc0678df4c736eeec50bc9fec37ed936a472a"

	// Busybox125Tag the 1.25 tag used for 1.25 busybox image
	Busybox125Tag = "1.25"

	// Busybox125Digest the digests used for 1.25 busybox image
	Busybox125Digest = "sha256:29f5d56d12684887bdfa50dcd29fc31eea4aaf4ad3bec43daf19026a7ce69912"

	// Busybox125ID the ID used for 1.25 busybox image
	Busybox125ID = "sha256:e02e811dd08fd49e7f6032625495118e63f597eb150403d02e3238af1df240ba"

	// HelloworldRepo the repository of hello-world image
	HelloworldRepo = "registry.hub.docker.com/library/hello-world"

	// HelloworldTag the tag used for hello-world image
	HelloworldTag = "linux"

	// HttpdRepo the repo for httpd image
	HttpdRepo = "registry.hub.docker.com/library/httpd"

	// HttpdTag the tag for httpd image
	HttpdTag = "2"

	// CniRepo the repo for cni image
	CniRepo = "calico/cni"

	// CniTag the tag for cni image
	CniTag = "v3.1.3"

	// GateWay default gateway for test
	GateWay = "192.168.1.1"

	// Subnet default subnet for test
	Subnet = "192.168.1.0/24"
)

// the following check funtions provide cgroup file avaible check
var (
	cgroupInfo *system.CgroupInfo

	// IsMemorySupport checks if memory cgroup is avaible
	IsMemorySupport = func() bool {
		return cgroupInfo.Memory.MemoryLimit
	}

	// IsMemorySwapSupport checks if memory swap cgroup is avaible
	IsMemorySwapSupport = func() bool {
		return cgroupInfo.Memory.MemorySwap
	}

	// IsMemorySwappinessSupport checks if memory swappiness cgroup is avaible
	IsMemorySwappinessSupport = func() bool {
		return cgroupInfo.Memory.MemorySwappiness
	}
)

func init() {
	cgroupInfo = system.NewCgroupInfo()
}

// GetBusybox gets image info from test environment variable.
func GetBusybox() {
	if env := os.Getenv("POUCH_BUSYBOX_REPO"); len(env) != 0 {
		BusyboxRepo = env
	}
	if env := os.Getenv("POUCH_BUSYBOX_TAG"); len(env) != 0 {
		BusyboxTag = env
	}
	if env := os.Getenv("POUCH_BUSYBOX_ID"); len(env) != 0 {
		BusyboxID = env
	}
	if env := os.Getenv("POUCH_BUSYBOX_DIGEST"); len(env) != 0 {
		BusyboxDigest = env
	}
	if env := os.Getenv("POUCH_BUSYBOX125_DIGEST"); len(env) != 0 {
		Busybox125Digest = env
	}
	if env := os.Getenv("POUCH_BUSYBOX125_ID"); len(env) != 0 {
		Busybox125ID = env
	}
}

// GetOtherImage gets other image info from test environment variable.
func GetOtherImage() {
	if env := os.Getenv("POUCH_HELLOWORLD_REPO"); len(env) != 0 {
		HelloworldRepo = env
	}
	if env := os.Getenv("POUCH_HELLOWORLD_TAG"); len(env) != 0 {
		HelloworldTag = env
	}
	if env := os.Getenv("POUCH_HTTPD_REPO"); len(env) != 0 {
		HttpdRepo = env
	}
	if env := os.Getenv("POUCH_HTTPD_TAG"); len(env) != 0 {
		HttpdTag = env
	}
	if env := os.Getenv("POUCH_CNI_REPO"); len(env) != 0 {
		CniRepo = env
	}
	if env := os.Getenv("POUCH_CNI_TAG"); len(env) != 0 {
		CniTag = env
	}
}

// GetTestNetwork get gateway and subnet from test environment variable.
func GetTestNetwork() {
	if len(os.Getenv("POUCH_TEST_GATEWAY")) != 0 {
		GateWay = os.Getenv("POUCH_TEST_GATEWAY")
	}
	if len(os.Getenv("POUCH_TEST_SUBNET")) != 0 {
		Subnet = os.Getenv("POUCH_TEST_SUBNET")
	}
}

// FindDisk finds a available disk, not partion
func FindDisk() (string, bool) {
	cmd := "lsblk -o NAME,TYPE -n | grep -w disk | head -1 | awk '{print $1}'"
	device := icmd.RunCommand("bash", "-c", cmd).Stdout()
	if device != "" {
		return strings.TrimSpace("/dev/" + device), true
	}
	return "", false
}

// IsLinux checks if the OS of test environment is Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsAliKernel checks if the kernel of test environment is AliKernel.
func IsAliKernel() bool {
	cmd := "uname -r | grep -i alios"
	return icmd.RunCommand("bash", "-c", cmd).ExitCode == 0
}

// IsDumbInitExist checks if the dumb-init binary exists on host.
func IsDumbInitExist() bool {
	if _, err := os.Stat("/usr/bin/dumb-init"); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// IsRuncVersionSupportRichContianer checks if the version of runc supports rich container.
func IsRuncVersionSupportRichContianer() bool {
	cmd := "runc -v|grep 1.0.0-rc4-1"
	return icmd.RunCommand("bash", "-c", cmd).ExitCode == 0
}

// IsHubConnected checks if hub address can be connected.
func IsHubConnected() bool {
	// TODO: found a proper way to test if hub address can be connected.
	return true
}

// IsDiskQuota checks if it can use disk quota for container.
func IsDiskQuota() bool {
	return icmd.RunCommand("which", "quotaon").ExitCode == 0
}

// IsPrjquota checks if there is prjquota set on test machine
func IsPrjquota() bool {
	return IsDiskQuota() &&
		(icmd.RunCommand("mount", "|grep prjquota").ExitCode == 0)
}

// IsGrpquota checks if there is grpquota set on test machine
func IsGrpquota() bool {
	return IsDiskQuota() &&
		(icmd.RunCommand("mount", "|grep grpquota").ExitCode == 0)
}

// IsLxcfsEnabled checks if the lxcfs is installed and service is enabled.
func IsLxcfsEnabled() bool {
	if icmd.RunCommand("which", "lxcfs").ExitCode != 0 {
		return false
	}
	if icmd.RunCommand("pgrep", "lxcfs").ExitCode != 0 {
		return false
	}
	cmd := "ps -ef |grep pouchd |grep \"enable\\-lxcfs\""
	return icmd.RunCommand("sh", "-c", cmd).ExitCode == 0
}

// IsCRIUExist checks if criu exist on machine.
func IsCRIUExist() bool {
	_, err := exec.LookPath("criu")
	return err == nil
}
