package spec

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func setupCap(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) error {
	//TODO setup capabilities
	return nil
}

func setupNs(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) error {
	//TODO setup ns
	return nil
}

func setupProcess(ctx context.Context, c *types.ContainerInfo, s *specs.Spec) (err error) {
	p := s.Process
	cmdArr := c.Config.Entrypoint
	if len(c.Config.Cmd) > 0 {
		cmdArr = append(cmdArr, c.Config.Cmd...)
	}
	if c.Config.Tty != nil {
		p.Terminal = *c.Config.Tty
	}
	p.Cwd = c.Config.WorkingDir
	p.Env = c.Config.Env
	p.Args = cmdArr
	if c.Config.User != "" {
		tmpArr := strings.SplitN(c.Config.User, ":", 2)
		var u, g string
		u = tmpArr[0]
		if len(tmpArr) == 2 {
			g = tmpArr[1]
		}
		user := &specs.User{}
		if uid, err := strconv.Atoi(u); err == nil {
			user.UID = uint32(uid)
		} else {
			user.Username = u
		}
		gid, err := strconv.Atoi(g)
		if err != nil || gid <= 0 {
			return fmt.Errorf("invalid gid: %d", gid)
		}
		user.GID = uint32(gid)
	}

	//TODO security config (including both seccomp and selinux)

	return nil
}
