package system

import (
	"bufio"
	"bytes"
	"net"
	"os/exec"
)

// GetNodeIP fetches node ip via command hostname.
// If it fails to get this, return empty string directly.
func GetNodeIP() string {
	output, err := exec.Command("hostname", "-i").CombinedOutput()
	if err != nil {
		return ""
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		ip := scanner.Text()
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}
