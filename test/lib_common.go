package main

import (
	"io/ioutil"
	"os/exec"

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

// runCmd runs Linux CMD and returns its stdout/stderr/error
func runCmd(cmd *exec.Cmd) ([]byte, []byte, error) {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	slurpErr, err := ioutil.ReadAll(stderr)
	defer stderr.Close()
	if err != nil {
		return nil, nil, err
	}

	slurpOut, err := ioutil.ReadAll(stdout)
	defer stdout.Close()
	if err != nil {
		return nil, slurpErr, err
	}

	if err := cmd.Wait(); err != nil {
		return slurpOut, slurpErr, err
	}

	return slurpOut, slurpErr, err
}
