package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gotestyourself/gotestyourself/icmd"
)

// WaitTimeout wait at most timeout nanoseconds,
// until the condition become true or timeout reached.
func WaitTimeout(timeout time.Duration, condition func() bool) bool {
	ch := make(chan bool, 1)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(timeout)
		done <- true
	}()

	for {
		select {
		case <-ch:
			return true
		case <-done:
			fmt.Printf("condition failed to return true within %f seconds.\n", timeout.Seconds())
			return false
		case <-ticker.C:
			if condition() {
				ch <- true
			}
		}
	}
}

// PartialEqual is used for check if string 'obtain' has substring 'expect',
// return can be used in check.Assert() or somewhere else.
func PartialEqual(obtain, expect string) error {
	if strings.Contains(obtain, expect) {
		return nil
	}

	return fmt.Errorf("obtained string: %s not include expected string: %s", obtain, expect)
}

// TrimAllSpaceAndNewline is used to strip all empty space and newline from a string.
func TrimAllSpaceAndNewline(input string) string {
	output := input
	for _, f := range []string{"\n", "\t", " "} {
		output = strings.Replace(output, f, "", -1)
	}

	return output
}

// GetMajMinNumOfDevice is used for getting major:minor device number
func GetMajMinNumOfDevice(device string) (string, bool) {
	cmd := fmt.Sprintf("lsblk -d -o MAJ:MIN %s | sed /MAJ:MIN/d | awk '{print $1}'", device)
	number := icmd.RunCommand("bash", "-c", cmd).Stdout()
	if number != "" {
		return strings.Trim(number, "\n"), true
	}
	return "", false
}

// StringSliceTrimSpace delete empty items from string slice
func StringSliceTrimSpace(input []string) ([]string, error) {
	output := []string{}

	for _, item := range input {
		str := strings.TrimSpace(item)
		if str != "" {
			output = append(output, str)
		}
	}

	return output, nil
}

// ParseCgroupFile parse cgroup path from cgroup file
func ParseCgroupFile(text string) map[string]string {
	cgroups := make(map[string]string)
	for _, t := range strings.Split(text, "\n") {
		parts := strings.SplitN(t, ":", 3)
		if len(parts) < 3 {
			continue
		}
		for _, sub := range strings.Split(parts[1], ",") {
			cgroups[sub] = parts[2]
		}
	}
	return cgroups
}

// FindCgroupMountpoint find cgroup mountpoint for a specified subsystem
func FindCgroupMountpoint(subsystem string) (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Fields(txt)
		if len(fields) < 5 {
			continue
		}
		if strings.Contains(txt, "cgroup") {
			for _, opt := range strings.Split(fields[len(fields)-1], ",") {
				if opt == subsystem {
					return fields[4], nil
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("failed to find %s cgroup mountpoint", subsystem)
}
