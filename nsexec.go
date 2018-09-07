// +build linux

package main

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/alibaba/pouch/pkg/nsexec"

	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/sirupsen/logrus"
)

// reexecRunNsExec will run nsenter in go init and join to container namespace
func reexecRunNsExec() {
	logrus.Infof("run reexecRunNsExec")
	initPipe := os.Getenv("_LIBCONTAINER_INITPIPE")
	pipeFd, err := strconv.Atoi(initPipe)
	if err != nil {
		panic("init pipe get failed")
	}

	f := os.NewFile(uintptr(pipeFd), "init")
	defer f.Close()

	nsexecOp := nsexec.Op{}
	err = json.NewDecoder(f).Decode(&nsexecOp)
	if err != nil {
		panic("op failed")
	}

	res := nsexec.Dispatch(nsexecOp)
	err = json.NewEncoder(f).Encode(res)
	if err != nil {
		panic("send data failed")
	}
}
