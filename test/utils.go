package main

import (
	"github.com/go-check/check"
)

// const defines common image name
const (
	busyboxImage = "registry.hub.docker.com/library/busybox:latest"
)

// VerifyCondition is used to check the condition value.
type VerifyCondition func() bool

// SkipIfFalse skips the suite, if any of the conditions is not satisfied.
func SkipIfFalse(c *check.C, conditions ...VerifyCondition) {
	for _, con := range conditions {
		if con() == false {
			c.Skip("Skip test as condition is not matched")
		}
	}
}
