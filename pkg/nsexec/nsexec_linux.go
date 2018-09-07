package nsexec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/utils"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink/nl"
)

// Package nsexec is to join container namespace to do some update in container such as update hostname, etc.
// Calls NsExec with param NsExecOp will do function in container namespace which has been registered
// by NsExecOp.OpCode. As to realize it, we set a new command nsexec which imports package
// github.com/opencontainers/runc/libcontainer/nsenter in which use C code to join namespace.

var (
	registerFuncMap = map[string]func(data []byte) (res []byte, err error){}
)

// Op is Nsexec input, opCode should be registered
type Op struct {
	OpCode string `json:"opCode"`
	Data   []byte `json:"data"`
}

// Result is the Nsexec result output
type Result struct {
	Finish bool   `json:"finish"`
	Data   []byte `json:"data"`
	ErrMsg string `json:"errMsg"`
}

type pid struct {
	Pid      int `json:"pid"`
	PidFirst int `json:"pid_first"`
}

type process struct {
	initArgs   []string
	stdin      io.Reader
	stdout     io.Writer
	stderr     io.Writer
	childPipe  *os.File
	parentPipe *os.File
	cmd        *exec.Cmd
}

const stdioFdCount = 3

// Register function which can execute in container namesapce
func Register(op string, handle func(data []byte) (res []byte, err error)) {
	if _, exist := registerFuncMap[op]; exist {
		panic(fmt.Sprintf("nsexec func already registered under name %s", op))
	}

	registerFuncMap[op] = handle
}

func newExecProcess(stdin io.Reader, stdout io.Writer, stderr io.Writer) *process {
	return &process{
		initArgs: []string{"/proc/self/exe", "nsexec"},
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
	}
}

func newNsExecCmd(p *process) {
	cmd := &exec.Cmd{
		Path: p.initArgs[0],
		Args: []string{p.initArgs[1]},
	}

	cmd.ExtraFiles = append(cmd.ExtraFiles, p.childPipe)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("_LIBCONTAINER_INITPIPE=%d", stdioFdCount+len(cmd.ExtraFiles)-1),
	)
	cmd.Stdin = p.stdin
	cmd.Stdout = p.stdout
	cmd.Stderr = p.stderr
	p.cmd = cmd
}

//setns is to wait init process
func setns(p *process) error {
	status, err := p.cmd.Process.Wait()
	if err != nil {
		p.cmd.Wait()
		return err
	}

	if !status.Success() {
		p.cmd.Wait()
		logrus.Errorf("in nsexec,set ns failed:%s", status.String())
		return &exec.ExitError{ProcessState: status}
	}

	var pid pid
	err = json.NewDecoder(p.parentPipe).Decode(&pid)
	if err != nil {
		p.cmd.Wait()
		return err
	}

	logrus.Infof("nsexec new pid is %d, pid_first is %d", pid.Pid, pid.PidFirst)

	pidFirstProcess, err := os.FindProcess(pid.PidFirst)
	if err != nil {
		return err
	}
	//wait for first child pid to prevent zombie process
	pidFirstProcess.Wait()

	childProcess, err := os.FindProcess(pid.Pid)
	if err != nil {
		return err
	}
	//set init process as cmd process
	p.cmd.Process = childProcess
	return nil
}

// can setns in order.
func orderNamespacePaths(namespaces map[configs.NamespaceType]string) ([]string, error) {
	paths := []string{}

	for _, ns := range configs.NamespaceTypes() {

		if p, ok := namespaces[ns]; ok && p != "" {
			// check if the requested namespace is supported
			if !configs.IsNamespaceSupported(ns) {
				return nil, fmt.Errorf("namespace %s is not supported", ns)
			}
			// only set to join this namespace if it exists
			if _, err := os.Lstat(p); err != nil {
				return nil, fmt.Errorf("running lstat on namespace path %q, error:%s", p, err.Error())
			}
			// do not allow namespace path with comma as we use it to separate
			// the namespace paths
			if strings.ContainsRune(p, ',') {
				return nil, fmt.Errorf("invalid path %s", p)
			}
			paths = append(paths, fmt.Sprintf("%s:%s", configs.NsName(ns), p))
		}

	}

	return paths, nil
}

// NsExec is the entry to do ns exec and then do something in container namespace
func NsExec(pid int, op Op) (result *Result, err error) {
	nsMap := map[configs.NamespaceType]string{}
	for _, nsType := range configs.NamespaceTypes() {
		if !configs.IsNamespaceSupported(nsType) {
			continue
		}

		//not set user namespace
		if nsType == configs.NEWUSER {
			continue
		}

		if _, ok := nsMap[nsType]; !ok {
			ns := configs.Namespace{Type: nsType}
			nsMap[ns.Type] = ns.GetPath(pid)
		}
	}

	// build bootstrap data
	// create the netlink message
	r := nl.NewNetlinkRequest(int(libcontainer.InitMsg), 0)

	// write cloneFlags
	r.AddData(&libcontainer.Int32msg{
		Type:  libcontainer.CloneFlagsAttr,
		Value: uint32(0),
	})

	// write custom namespace paths
	if len(nsMap) > 0 {
		nsPaths, err := orderNamespacePaths(nsMap)
		if err != nil {
			return nil, err
		}
		r.AddData(&libcontainer.Bytemsg{
			Type:  libcontainer.NsPathsAttr,
			Value: []byte(strings.Join(nsPaths, ",")),
		})
	}

	bootstrapData := bytes.NewReader(r.Serialize())

	//create peer pipe
	parentPipe, childPipe, err := utils.NewSockPair("init")
	if err != nil {
		return nil, err
	}

	defer parentPipe.Close()

	buffer := bytes.NewBuffer(nil)
	p := newExecProcess(nil, nil, buffer)
	p.childPipe = childPipe
	p.parentPipe = parentPipe

	newNsExecCmd(p)
	err = p.cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("start nsexec failed:%s", err.Error())
	}

	childPipe.Close()

	logrus.Infof("child pid is %d", p.cmd.Process.Pid)

	defer func() {
		if err != nil {
			logrus.Errorf("nsexec read stderr from child:%s", buffer.String())
		}
	}()

	//send bootstrapData to nsenter to join namespace
	_, err = io.Copy(parentPipe, bootstrapData)
	if err != nil {
		return nil, fmt.Errorf("send bootstrapData to nsexec failed:%s", err.Error())
	}

	if err = setns(p); err != nil {
		return nil, fmt.Errorf("set exec ns error:%s", err.Error())
	}

	if err = json.NewEncoder(parentPipe).Encode(op); err != nil {
		return nil, fmt.Errorf("send NsExecOp error:%s", err.Error())
	}

	res := &Result{}
	if err = json.NewDecoder(parentPipe).Decode(res); err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("read res from pipe error:%s", err.Error())
		}
	}

	logrus.Infof("NsExecResult: %v", res)

	if !res.Finish {
		return nil, fmt.Errorf("not finish nsexec")
	}

	if err := syscall.Shutdown(int(p.parentPipe.Fd()), syscall.SHUT_WR); err != nil {
		return nil, fmt.Errorf("shutting down init pipe failed:%s", err.Error())
	}
	// Must be done after Shutdown so the child will exit and we can wait for it.
	p.cmd.Wait()
	return res, nil
}
