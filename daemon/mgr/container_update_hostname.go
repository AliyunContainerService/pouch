package mgr

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"syscall"

	"github.com/alibaba/pouch/pkg/nsexec"
)

func init() {
	nsexec.Register("sethostname", handleUpdateHostname)
}

func handleUpdateHostname(data []byte) (res []byte, err error) {
	oldHostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	err = syscall.Sethostname(data)
	if err != nil {
		return nil, err
	}

	//update /etc/hostname and /etc/hosts
	err = ioutil.WriteFile("/etc/hostname", append(data, '\n'), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to update /etc/hostname: %v", err)
	}

	hosts, err := ioutil.ReadFile("/etc/hosts")
	if err != nil {
		return nil, fmt.Errorf("failed to update /etc/hosts: %v", err)
	}

	exp1 := regexp.MustCompile(fmt.Sprintf(`\s%s\s`, oldHostname))
	exp2 := regexp.MustCompile(fmt.Sprintf(`\s%s$`, oldHostname))

	replaceStr1 := fmt.Sprintf(" %s ", string(data))
	replaceStr2 := fmt.Sprintf(" %s", string(data))

	d := string(hosts)
	lines := strings.Split(d, "\n")
	for i, line := range lines {
		if !strings.Contains(line, oldHostname) {
			continue
		}

		newLine := line

		if exp1.FindString(newLine) != "" {
			newLine = exp1.ReplaceAllString(newLine, replaceStr1)
		}

		if exp2.FindString(newLine) != "" {
			newLine = exp2.ReplaceAllString(newLine, replaceStr2)
		}

		lines[i] = newLine
	}

	newHosts := strings.Join(lines, "\n")
	err = ioutil.WriteFile("/etc/hosts", []byte(newHosts), 0644)

	return nil, err
}

func updateHostnameForRunningContainer(hostname string, container *Container) error {
	pid := container.State.Pid
	if pid <= 0 {
		return fmt.Errorf("container pid %d is not vaild", pid)
	}

	_, err := nsexec.NsExec(int(pid), nsexec.Op{OpCode: "sethostname", Data: []byte(hostname)})
	if err != nil {
		return err
	}

	return nil
}
