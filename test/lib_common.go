package main

import (
	"io/ioutil"
	"os/exec"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
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

// PouchCmd contains pouch commandline and expected result
type PouchCmd struct {
	// pouch binary locaiton, by default it is "/usr/bin/pouch"
	binary string

	// pouch argument
	args []string

	// MustRequired: result mean whether or not the CMD is expected to succeed
	result bool

	// return value of the CMD, MustRequired if "result == false"
	returnValue int

	// expected contained output
	outContains string

	// expected stdout
	stdout string

	// expected stderr
	stderr string
}

// RunCmd executes CMD, checks return value and stdout/stderr according to user's config
func RunCmd(c *check.C, cmd *PouchCmd) {

	if cmd.binary == "" {
		cmd.binary = defaultBinary
	}

	ret := icmd.RunCmd(icmd.Command(cmd.binary, cmd.args...))

	// check result
	if cmd.result == true {
		c.Assert(ret.ExitCode, check.Equals, 0,
			check.Commentf("[FAIL]: CMD[%s %s] return %d,\n"+
				"Stdout: %s\n Stderr: %s",
				cmd.binary, cmd.args, ret.ExitCode, ret.Stdout(), ret.Stderr()))

		if cmd.outContains != "" {
			// Add "(?s)" to match "\n", add ".*" to return success as lone as ret.Combined() contains cmd.outContains
			c.Assert(ret.Combined(), check.Matches, "(?s).*"+cmd.outContains+".*",
				check.Commentf("\n[FAIL]: CMD[%s %s] return %d,\n"+
					"Stdout: \n%s\nStderr:\n%s\n",
					cmd.binary, cmd.args, ret.ExitCode, ret.Stdout(), ret.Stderr()))
		}

		if cmd.stdout != "" {
			c.Assert(ret.Stdout(), check.Matches, cmd.stdout,
				check.Commentf("\n[FAIL]: CMD[%s %s] return %d,\n"+
					"Stdout: \n%s\nStderr:\n%s\n",
					cmd.binary, cmd.args, ret.ExitCode, ret.Stdout(), ret.Stderr()))
		}
	} else {
		c.Assert(ret.ExitCode, check.Equals, ret.ExitCode,
			check.Commentf("[FAIL]: CMD[%s %s] return %d,\n"+
				"Stdout: %s\n Stderr: %s",
				cmd.binary, cmd.args, ret.ExitCode, ret.Stdout(), ret.Stderr()))

		if cmd.outContains != "" {
			// Add "(?s)" to match "\n", add ".*" to return success as lone as ret.Combined() contains cmd.outContains
			c.Assert(ret.Combined(), check.Matches, "(?s).*"+cmd.outContains+".*",
				check.Commentf("\n[FAIL]: CMD[%s %s] return %d,\n"+
					"Stdout: \n%s\nStderr:\n%s\n",
					cmd.binary, cmd.args, ret.ExitCode, ret.Stdout(), ret.Stderr()))
		}

		if cmd.stdout != "" {
			c.Assert(ret.Stdout(), check.Matches, cmd.stdout,
				check.Commentf("\n[FAIL]: CMD[%s %s] return %d,\n"+
					"Stdout: \n%s\nStderr:\n%s\n",
					cmd.binary, cmd.args, ret.ExitCode, ret.Stdout(), ret.Stderr()))

		}

		if cmd.stderr != "" {

			c.Assert(ret.Stderr(), check.Matches, cmd.stderr,
				check.Commentf("\n[FAIL]: CMD[%s %s] return %d,\n"+
					"Stdout: \n%s\nStderr:\n%s\n",
					cmd.binary, cmd.args, ret.ExitCode, ret.Stdout(), ret.Stderr()))
		}

	}
}
