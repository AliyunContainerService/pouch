package util

import (
	"fmt"
	"time"
)

// WaitTimeout wait at most timeout nanoseconds,
// until the conditon become true or timeout reached.
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
