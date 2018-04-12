package system

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// file to check to determine Operating System
const etcOsRelease = "/etc/os-release"

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
