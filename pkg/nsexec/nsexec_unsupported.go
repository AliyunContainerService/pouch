// +build !linux

package nsexec

import "errors"

//nsexec op struct, opCode should be registered
type NsExecOp struct {
	OpCode string `json:"opCode"`
	Data   []byte `json:"data"`
}

type NsExecResult struct {
	Finish bool   `json:"finish"`
	Data   []byte `json:"data"`
	ErrMsg string `json:"errMsg"`
}

func NsExec(pid int, op NsExecOp) (result *NsExecResult, err error) {
	return errors.New("not supported")
}
