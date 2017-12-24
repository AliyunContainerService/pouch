package command

import (
	"github.com/alibaba/pouch/test/environment"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRun runs the specific pouch command and return result.
func PouchRun(args ...string) *icmd.Result {
	return icmd.RunCmd(PouchCmd(args...))
}

// PouchCmd will return default pouch binary command with args.
func PouchCmd(args ...string) icmd.Cmd {
	return icmd.Cmd{Command: append([]string{environment.PouchBinary}, args...)}
}
