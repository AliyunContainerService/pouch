package nsexec

import "fmt"

// Dispatch should be called by pouchd nsexec to do function in container namespace
func Dispatch(op Op) *Result {
	f, exist := registerFuncMap[op.OpCode]
	if !exist {
		return &Result{
			Finish: true,
			ErrMsg: fmt.Sprintf("not register nsexec op %s", op.OpCode),
		}
	}

	res, err := f(op.Data)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	return &Result{
		Finish: true,
		ErrMsg: errMsg,
		Data:   res,
	}
}
