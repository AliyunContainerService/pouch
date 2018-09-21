package mgr

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/opts"
	"github.com/alibaba/pouch/apis/types"
	networktypes "github.com/alibaba/pouch/network/types"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/randomid"

	"github.com/opencontainers/selinux/go-selinux/label"
	"github.com/pkg/errors"
)

// containerID returns the container's id, the parameter 'nameOrPrefix' may be container's
// name, id or prefix id.
func (mgr *ContainerManager) containerID(nameOrPrefix string) (string, error) {
	var obj meta.Object

	// name is the container's name.
	id, ok := mgr.NameToID.Get(nameOrPrefix).String()
	if ok {
		return id, nil
	}

	// name is the container's prefix of the id.
	objs, err := mgr.Store.GetWithPrefix(nameOrPrefix)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get container info with prefix %s", nameOrPrefix)
	}
	if len(objs) > 1 {
		return "", errors.Wrapf(errtypes.ErrTooMany, "container %s", nameOrPrefix)
	}
	if len(objs) == 0 {
		return "", errors.Wrapf(errtypes.ErrNotfound, "container %s", nameOrPrefix)
	}
	obj = objs[0]

	con, ok := obj.(*Container)
	if !ok {
		return "", fmt.Errorf("failed to get container info, invalid meta's type")
	}

	return con.ID, nil
}

func (mgr *ContainerManager) container(nameOrPrefix string) (*Container, error) {
	id, err := mgr.containerID(nameOrPrefix)
	if err != nil {
		return nil, err
	}

	// lookup again
	res, ok := mgr.cache.Get(id).Result()
	if ok {
		return res.(*Container), nil
	}

	return nil, errors.Wrapf(errtypes.ErrNotfound, "container %s", nameOrPrefix)
}

// generateID generates an ID for newly created container. We must ensure that
// this ID has not used yet.
func (mgr *ContainerManager) generateID() (string, error) {
	var id string
	for {
		id = randomid.Generate()
		_, err := mgr.Store.Get(id)
		if err != nil {
			if merr, ok := err.(meta.Error); ok && merr.IsNotfound() {
				break
			}
			return "", err
		}
	}
	return id, nil
}

// generateName generates container name by container ID.
// It get first 6 continuous letters which has not been taken.
// TODO: take a shorter than 6 letters ID into consideration.
// FIXME: there is possibility that for block loops forever.
func (mgr *ContainerManager) generateName(id string) string {
	var name string
	i := 0
	for {
		if i+6 > len(id) {
			break
		}
		name = id[i : i+6]
		i++
		if !mgr.NameToID.Get(name).Exist() {
			break
		}
	}
	return name
}

// getRuntime returns runtime real path.
// TODO(huamin): do we need validate runtime is executable ?
func (mgr *ContainerManager) getRuntime(runtime string) (string, error) {
	r, exist := mgr.Config.Runtimes[runtime]
	if !exist {
		return "", fmt.Errorf("failed to find runtime %s in daemon config", runtime)
	}

	// it is ok to use runtime name as a path.
	rPath := runtime
	// generally speaking, path is not be empty, but we not forbid empty path
	// in config set, since name can be a path too.
	if r.Path != "" {
		rPath = r.Path
	}

	// if Runtime has args, use script path as runtime path.
	if len(r.RuntimeArgs) > 0 {
		rPath = filepath.Join(mgr.Config.HomeDir, RuntimeDir, runtime)
	}

	return rPath, nil
}

// BuildContainerEndpoint is used to build container's endpoint config.
func BuildContainerEndpoint(c *Container) *networktypes.Endpoint {
	return &networktypes.Endpoint{
		Owner:           c.ID,
		Hostname:        c.Config.Hostname,
		Domainname:      c.Config.Domainname,
		HostsPath:       c.HostsPath,
		ExtraHosts:      c.HostConfig.ExtraHosts,
		HostnamePath:    c.HostnamePath,
		ResolvConfPath:  c.ResolvConfPath,
		NetworkDisabled: c.Config.NetworkDisabled,
		NetworkMode:     c.HostConfig.NetworkMode,
		DNS:             c.HostConfig.DNS,
		DNSOptions:      c.HostConfig.DNSOptions,
		DNSSearch:       c.HostConfig.DNSSearch,
		MacAddress:      c.Config.MacAddress,
		PublishAllPorts: c.HostConfig.PublishAllPorts,
		ExposedPorts:    c.Config.ExposedPorts,
		PortBindings:    c.HostConfig.PortBindings,
		NetworkConfig:   c.NetworkSettings,
	}
}

func parseSecurityOpts(c *Container, securityOpts []string) error {
	var (
		labelOpts []string
		err       error
	)
	for _, securityOpt := range securityOpts {
		if securityOpt == "no-new-privileges" {
			c.NoNewPrivileges = true
			continue
		}
		fields := strings.SplitN(securityOpt, "=", 2)
		if len(fields) != 2 {
			return fmt.Errorf("invalid --security-opt %s: must be in format of key=value", securityOpt)
		}
		key, value := fields[0], fields[1]
		switch key {
		// TODO: handle other security options.
		case "apparmor":
			c.AppArmorProfile = value
		case "seccomp":
			c.SeccompProfile = value
		case "label":
			labelOpts = append(labelOpts, value)
		default:
			return fmt.Errorf("invalid type %s in --security-opt %s: unknown type from apparmor, seccomp, no-new-privileges and SELinux label", key, securityOpt)
		}
	}

	if len(labelOpts) == 0 {
		return nil
	}
	c.ProcessLabel, c.MountLabel, err = label.InitLabels(labelOpts)
	if err != nil {
		return fmt.Errorf("failed to init labels: %v", err)
	}

	return nil
}

// fieldsASCII is similar to strings.Fields but only allows ASCII whitespaces
func fieldsASCII(s string) []string {
	fn := func(r rune) bool {
		switch r {
		case '\t', '\n', '\f', '\r', ' ':
			return true
		}
		return false
	}
	return strings.FieldsFunc(s, fn)
}

func parsePSOutput(output []byte, pids []int) (*types.ContainerProcessList, error) {
	procList := &types.ContainerProcessList{}

	lines := strings.Split(string(output), "\n")
	procList.Titles = fieldsASCII(lines[0])

	pidIndex := -1
	for i, name := range procList.Titles {
		if name == "PID" {
			pidIndex = i
		}
	}
	if pidIndex == -1 {
		return nil, fmt.Errorf("Couldn't find PID field in ps output")
	}

	// loop through the output and extract the PID from each line
	for _, line := range lines[1:] {
		if len(line) == 0 {
			continue
		}
		fields := fieldsASCII(line)
		p, err := strconv.Atoi(fields[pidIndex])
		if err != nil {
			return nil, fmt.Errorf("Unexpected pid '%s': %s", fields[pidIndex], err)
		}

		for _, pid := range pids {
			if pid == p {
				// Make sure number of fields equals number of header titles
				// merging "overhanging" fields
				process := fields[:len(procList.Titles)-1]
				process = append(process, strings.Join(fields[len(procList.Titles)-1:], " "))
				procList.Processes = append(procList.Processes, process)
			}
		}
	}
	return procList, nil
}

// amendContainerSettings modify config settings to wanted,
// it will be call before container created.
func amendContainerSettings(config *types.ContainerConfig, hostConfig *types.HostConfig) {
	r := &hostConfig.Resources
	if r.Memory > 0 && r.MemorySwap == 0 {
		r.MemorySwap = 2 * r.Memory
	}
}

func mergeEnvSlice(newEnv, oldEnv []string) ([]string, error) {
	// if newEnv is empty, return old env slice
	if len(newEnv) == 0 {
		return oldEnv, nil
	}

	newEnvMap, err := opts.ParseEnv(newEnv)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse new env")
	}

	oldEnvMap, err := opts.ParseEnv(oldEnv)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse old env")
	}

	for k, v := range newEnvMap {
		// key should not be empty
		if k == "" {
			continue
		}

		// add or change an env
		if v != "" {
			oldEnvMap[k] = v
			continue
		}

		// value is empty, we need delete the env
		if _, exists := oldEnvMap[k]; exists {
			delete(oldEnvMap, k)
		}
	}

	newEnvSlice := []string{}
	for k, v := range oldEnvMap {
		newEnvSlice = append(newEnvSlice, fmt.Sprintf("%s=%s", k, v))
	}

	return newEnvSlice, nil
}
