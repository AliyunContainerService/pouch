package util

import (
	"fmt"
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
			if condition() == true {
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
