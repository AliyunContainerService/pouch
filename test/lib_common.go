package main

import (
	"github.com/go-check/check"
)

// VerifyCondition is used to check whether or not
// the test should be skipped.
type VerifyCondition func() bool

// SkipIfFalse skips the test/test suite, if any of the
// conditions is not satisfied.
func SkipIfFalse(c *check.C, conditions ...VerifyCondition) {
	for _, con := range conditions {
		ret := con()
		if ret == false {
			c.Skip("Skip test as condition is not matched")
		}
	}
}
