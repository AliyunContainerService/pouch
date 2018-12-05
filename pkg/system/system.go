package system

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// file to check to determine Operating System
const etcOsRelease = "/etc/os-release"

// Info defines system info on current machine
type Info struct {
	AppArmor bool
	Seccomp  bool

	*CgroupInfo
}

// NewInfo creates a system info about current machine.
func NewInfo() *Info {
	info := &Info{CgroupInfo: NewCgroupInfo()}

	// Check if AppArmor is supported.
	// isAppArmorEnabled returns true if apparmor is enabled for the host.
	// This function is forked from
	// https://github.com/opencontainers/runc/blob/1a81e9ab1f138c091fe5c86d0883f87716088527/libcontainer/apparmor/apparmor.go
	// to avoid the libapparmor dependency.
	if _, err := os.Stat("/sys/kernel/security/apparmor"); err == nil && os.Getenv("container") == "" {
		if _, err = os.Stat("/sbin/apparmor_parser"); err == nil {
			buf, err := ioutil.ReadFile("/sys/module/apparmor/parameters/enabled")
			if err == nil && len(buf) > 1 && buf[0] == 'Y' {
				info.AppArmor = true
			}
		}
	}

	// Check if Seccomp is supported, via CONFIG_SECCOMP.
	if err := unix.Prctl(unix.PR_GET_SECCOMP, 0, 0, 0, 0); err != unix.EINVAL {
		// Make sure the kernel has CONFIG_SECCOMP_FILTER.
		if err := unix.Prctl(unix.PR_SET_SECCOMP, unix.SECCOMP_MODE_FILTER, 0, 0, 0); err != unix.EINVAL {
			info.Seccomp = true
		}
	}

	return info
}

// getSysInfo gets sysinfo.
func getSysInfo() (*syscall.Sysinfo_t, error) {
	si := &syscall.Sysinfo_t{}
	err := syscall.Sysinfo(si)
	if err != nil {
		return nil, err
	}
	return si, nil
}

// GetTotalMem gets total ram of host.
func GetTotalMem() (uint64, error) {
	si, err := getSysInfo()
	if err != nil {
		return 0, err
	}
	return si.Totalram, nil
}

// GetDevID returns device id via syscall according to the input directory.
func GetDevID(dir string) (uint64, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(dir, &st); err != nil {
		return 0, errors.Wrapf(err, "failed to get device id of directory: (%s)", dir)
	}
	return st.Dev, nil
}

// GetSerialNumber gets serial number or a machine.
func GetSerialNumber() string {
	var sn string
	if b, e := exec.Command("dmidecode", "-s", "system-serial-number").CombinedOutput(); e == nil {
		scanner := bufio.NewScanner(bytes.NewReader(b))
		for scanner.Scan() {
			sn = scanner.Text()
		}
	}
	if len(strings.Fields(sn)) != 0 {
		sn = strings.Fields(sn)[0]
	}
	for i := 0; i < 10; i++ {
		if _, ex := os.Stat("/usr/alisys/dragoon/libexec/armory/bin/armoryinfo"); ex == nil {
			if b, e := exec.Command("/usr/alisys/dragoon/libexec/armory/bin/armoryinfo", "sn").CombinedOutput(); e == nil {
				sn = strings.TrimSpace(string(b))
			}
		}
		if sn != "" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return sn
}

// GetNodeIP fetches node ip via command hostname.
// If it fails to get this, return empty string directly.
func GetNodeIP() string {
	output, err := exec.Command("hostname", "-i").CombinedOutput()
	if err != nil {
		return ""
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		ip := scanner.Text()
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}

// GetOSName gets data in /etc/os-release and gets OS name.
// For example, in a Ubuntu host, fetched data are like:
// root@i-8brpbc9t:~# cat /etc/os-release
// NAME="Ubuntu"
// VERSION="16.04.2 LTS (Xenial Xerus)"
// ID=ubuntu
// ID_LIKE=debian
// PRETTY_NAME="Ubuntu 16.04.2 LTS"
// VERSION_ID="16.04"
// HOME_URL="http://www.ubuntu.com/"
// SUPPORT_URL="http://help.ubuntu.com/"
// BUG_REPORT_URL="http://bugs.launchpad.net/ubuntu/"
// VERSION_CODENAME=xenial
// UBUNTU_CODENAME=xenial
func GetOSName() (string, error) {
	etcOsReleaseFile, err := os.Open(etcOsRelease)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to open %s: %v", etcOsRelease, err)
		}
	}
	defer etcOsReleaseFile.Close()

	var prettyName string

	scanner := bufio.NewScanner(etcOsReleaseFile)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "PRETTY_NAME=") {
			continue
		}

		data := strings.SplitN(line, "=", 2)
		prettyName = data[1]
		return prettyName, nil
	}

	return "Linux", nil

}
