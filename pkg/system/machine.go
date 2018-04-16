package system

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"
	"time"
)

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
