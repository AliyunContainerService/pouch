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
func runCmd(cmd *exec.Cmd) (string, string, error) {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}

	if err := cmd.Start(); err != nil {
		return "", "", err
	}

	slurpErr, err := ioutil.ReadAll(stderr)
	defer stderr.Close()
	if err != nil {
		return "", "", err
	}

	slurpOut, err := ioutil.ReadAll(stdout)
	defer stdout.Close()
	if err != nil {
		return "", string(slurpErr[:]), err
	}

	if err := cmd.Wait(); err != nil {
		return string(slurpOut[:]), string(slurpErr[:]), err
	}

	return string(slurpOut[:]), string(slurpErr[:]), err
}

// runCmdPos asserts failure when CMD returns error
func runCmdPos(c *check.C, cmd *exec.Cmd) {
	stdout, stderr, err := runCmd(cmd)
	c.Assert(err, check.IsNil, check.Commentf("failed to run CMD: [ %s ] \n OUT: [ %s ] \n ERR: [ %s ]", cmd, stdout, stderr))
}

// runCmdNeg asserts failure when CMD does not return error
func runCmdNeg(c *check.C, cmd *exec.Cmd) {
	stdout, stderr, err := runCmd(cmd)
	c.Assert(err, check.NotNil, check.Commentf("failed to run CMD: [ %s ] \n OUT: [ %s ] \n ERR: [ %s ]", cmd, stdout, stderr))
}
