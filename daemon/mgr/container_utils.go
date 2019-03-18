package mgr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/opts"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"
	networktypes "github.com/alibaba/pouch/network/types"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/randomid"

	"github.com/containerd/containerd/runtime/linux/runctypes"
	runcoptions "github.com/containerd/containerd/runtime/v2/runc/options"
	specs "github.com/opencontainers/runtime-spec/specs-go"
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

// getRuntimeType returns containerd runtime type, type shim v1 by default.
func (mgr *ContainerManager) getRuntimeType(runtime string) (string, error) {
	r, exist := mgr.Config.Runtimes[runtime]
	if !exist {
		return "", fmt.Errorf("failed to find runtime %s in daemon config", runtime)
	}
	if r.Type == "" {
		return ctrd.RuntimeTypeV1, nil
	}
	return r.Type, nil
}

// generateRuntimeOptions generate options from daemon runtime configurations.
func (mgr *ContainerManager) generateRuntimeOptions(runtime string) (interface{}, error) {
	r, exist := mgr.Config.Runtimes[runtime]
	if !exist {
		return nil, fmt.Errorf("failed to find runtime %s in daemon config", runtime)
	}

	var options interface{}
	switch o := r.Options.(type) {
	// io.containerd.runtime.v1.linux
	case *runctypes.RuncOptions:
		options = &runctypes.RuncOptions{
			Runtime:       r.Path,
			RuntimeRoot:   ctrd.RuntimeRoot,
			CriuPath:      o.CriuPath,
			SystemdCgroup: mgr.Config.UseSystemd(),
		}
	// io.containerd.runc.v1
	case *runcoptions.Options:
		options = &runcoptions.Options{
			NoPivotRoot:   o.NoPivotRoot,
			NoNewKeyring:  o.NoNewKeyring,
			ShimCgroup:    o.ShimCgroup,
			IoUid:         o.IoUid,
			IoGid:         o.IoGid,
			BinaryName:    r.Path,
			Root:          ctrd.RuntimeRoot,
			CriuPath:      o.CriuPath,
			SystemdCgroup: mgr.Config.UseSystemd(),
		}
	// TODO: support other v2 shim options.
	default:
		return nil, nil
	}

	return options, nil
}

// getContainerSpec returns container runtime spec, unmarshal spec from config.json
func (mgr *ContainerManager) getContainerSpec(c *Container) (*specs.Spec, error) {
	configFile := filepath.Join(path.Dir(c.BaseFS), "config.json")
	var spec specs.Spec
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
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

// mergeEnvSlice merges two parts into a singe one.
// Here are some cases:
// 1. container creation needs to merge user input envs and envs inherited from image;
// 2. update action with env needs to merge original envs and the user input envs;
// 3. exec action needs to merge container's original envs and user input envs.
func mergeEnvSlice(new, old []string) ([]string, error) {
	// if newEnv is empty, return old env slice
	if len(new) == 0 {
		return old, nil
	}

	newEnvs, err := opts.ParseEnvs(new)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse new env")
	}

	// TODO: do we need to valid the old one?
	oldEnvs, err := opts.ParseEnvs(old)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse old env")
	}

	oldEnvsMap := opts.ValidSliceEnvsToMap(oldEnvs)

	for _, env := range newEnvs {
		arr := strings.SplitN(env, "=", 2)
		if len(arr) == 1 {
			// there are two cases, the first is 'KEY=', the second is just 'KEY'
			if len(env) == len(arr[0]) {
				// the case of 'KEY'. It is valid to remove the env from original one.
				if _, exits := oldEnvsMap[arr[0]]; exits {
					delete(oldEnvsMap, arr[0])
				}
			} else {
				// the case of 'KEY='. It is valid to set env value empty.
				oldEnvsMap[arr[0]] = ""
			}
		} else {
			oldEnvsMap[arr[0]] = arr[1]
		}
	}

	return opts.ValidMapEnvsToSlice(oldEnvsMap), nil
}

func mergeAnnotation(newAnnotation, oldAnnotation map[string]string) map[string]string {
	if len(newAnnotation) == 0 {
		return oldAnnotation
	}

	if len(oldAnnotation) == 0 {
		oldAnnotation = make(map[string]string)
	}

	for k, v := range newAnnotation {
		oldAnnotation[k] = v
	}

	return oldAnnotation
}
