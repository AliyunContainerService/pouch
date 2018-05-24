package environment

import (
	"os"
	"runtime"
	"strings"

	"github.com/alibaba/pouch/client"

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

	// BusyboxID the digest ID used for busybox image
	BusyboxID = "sha256:8ac48589692a53a9b8c2d1ceaa6b402665aa7fe667ba51ccc03002300856d8c7"

	// BusyboxTag the tag used for busybox image
	BusyboxTag = "1.28"

	// BusyboxDigest the digest used for busybox image
	BusyboxDigest = "sha256:58ac43b2cc92c687a32c8be6278e50a063579655fe3090125dcb2af0ff9e1a64"

	// HelloworldRepo the repository of hello-world image
	HelloworldRepo = "registry.hub.docker.com/library/hello-world"

	// HelloworldTag the tag used for hello-world image
	HelloworldTag = "linux"

	// GateWay default gateway for test
	GateWay = "192.168.1.1"

	// Subnet default subnet for test
	Subnet = "192.168.1.0/24"
)

// GetBusybox get image info from test environment variable.
func GetBusybox() {
	if len(os.Getenv("POUCH_BUSYBOXREPO")) != 0 {
		BusyboxRepo = os.Getenv("POUCH_BUSYBOXREPO")
	}
	if len(os.Getenv("POUCH_BUSYBOXTAG")) != 0 {
		BusyboxTag = os.Getenv("POUCH_BUSYBOXTAG")
	}
}

// GetHelloWorld get image info from test environment variable.
func GetHelloWorld() {
	if len(os.Getenv("POUCH_HELLOWORLDREPO")) != 0 {
		HelloworldRepo = os.Getenv("POUCH_HELLOWORLDREPO")
	}
	if len(os.Getenv("POUCH_HELLOWORLDTAG")) != 0 {
		HelloworldTag = os.Getenv("POUCH_HELLOWORLDTAG")
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
	if icmd.RunCommand("bash", "-c", cmd).ExitCode == 0 {
		return true
	}
	return false
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
	if icmd.RunCommand("bash", "-c", cmd).ExitCode == 0 {
		return true
	}
	return false
}

// IsHubConnected checks if hub address can be connected.
func IsHubConnected() bool {
	// TODO: found a proper way to test if hub address can be connected.
	return true
}

// IsDiskQuota checks if it can use disk quota for container.
func IsDiskQuota() bool {
	if icmd.RunCommand("which", "quotaon").ExitCode == 0 {
		return true
	}
	return false
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
	if icmd.RunCommand("sh", "-c", cmd).ExitCode != 0 {
		return false
	}
	return true
}
