package mgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/pkg/errors"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

const (
	// RPrivatePropagationMode represents mount propagation rprivate.
	RPrivatePropagationMode = "rprivate"
	// PrivatePropagationMode represents mount propagation private.
	PrivatePropagationMode = "private"
	// RSharedPropagationMode represents mount propagation rshared.
	RSharedPropagationMode = "rshared"
	// SharedPropagationMode represents mount propagation shared.
	SharedPropagationMode = "shared"
	// RSlavePropagationMode represents mount propagation rslave.
	RSlavePropagationMode = "rslave"
	// SlavePropagationMode represents mount propagation slave.
	SlavePropagationMode = "slave"
)

func clearReadonly(m *specs.Mount) {
	var opts []string
	for _, o := range m.Options {
		if o != "ro" {
			opts = append(opts, o)
		}
	}
	m.Options = opts
}

func overrideDefaultMount(mounts []specs.Mount, c *Container, s *specs.Spec) ([]specs.Mount, error) {
	for _, sm := range s.Mounts {
		dup := false
		for _, cm := range c.Mounts {
			if sm.Destination == cm.Destination {
				dup = true
				break
			}
		}
		if dup {
			continue
		}

		mounts = append(mounts, sm)
	}

	return mounts, nil
}

func mergeContainerMount(mounts []specs.Mount, c *Container, s *specs.Spec) ([]specs.Mount, error) {
	for _, mp := range c.Mounts {
		if trySetupNetworkMount(mp, c) {
			// ignore the network mount, we will handle it later.
			continue
		}

		// check duplicate mountpoint
		for _, sm := range mounts {
			if sm.Destination == mp.Destination {
				return nil, fmt.Errorf("duplicate mount point: %s", mp.Destination)
			}
		}

		pg := mp.Propagation
		rootfspg := s.Linux.RootfsPropagation
		// Set rootfs propagation, default setting is private.
		switch pg {
		case SharedPropagationMode, RSharedPropagationMode:
			if rootfspg != SharedPropagationMode && rootfspg != RSharedPropagationMode {
				s.Linux.RootfsPropagation = SharedPropagationMode
			}
		case SlavePropagationMode, RSlavePropagationMode:
			if rootfspg != SharedPropagationMode && rootfspg != RSharedPropagationMode &&
				rootfspg != SlavePropagationMode && rootfspg != RSlavePropagationMode {
				s.Linux.RootfsPropagation = RSlavePropagationMode
			}
		}

		opts := []string{"rbind"}
		if !mp.RW {
			opts = append(opts, "ro")
		}

		// set rprivate propagation to bind mount if pg is ""
		if pg == "" {
			pg = RPrivatePropagationMode
		}
		opts = append(opts, pg)

		// TODO: support copy data.

		mounts = append(mounts, specs.Mount{
			Source:      mp.Source,
			Destination: mp.Destination,
			Type:        "bind",
			Options:     opts,
		})
	}

	// if disable hostfiles, we will not mount the hosts files into container.
	if !c.Config.DisableNetworkFiles {
		mounts = append(mounts, generateNetworkMounts(c)...)
	}

	return mounts, nil
}

// setupMounts create mount spec.
func setupMounts(ctx context.Context, c *Container, s *specs.Spec) error {
	var (
		mounts []specs.Mount
		err    error
	)

	// Override the default mounts which are duplicate with user defined ones.
	mounts, err = overrideDefaultMount(mounts, c, s)
	if err != nil {
		return errors.Wrap(err, "failed to override default spec mounts")
	}

	// user defined mount
	mounts, err = mergeContainerMount(mounts, c, s)
	if err != nil {
		return errors.Wrap(err, "failed to merge container mounts")
	}

	// modify share memory size, and change rw mode for privileged mode.
	for i := range mounts {
		if mounts[i].Destination == "/dev/shm" && c.HostConfig.ShmSize != nil &&
			*c.HostConfig.ShmSize != 0 {
			for idx, v := range mounts[i].Options {
				if strings.Contains(v, "size=") {
					mounts[i].Options[idx] = fmt.Sprintf("size=%s",
						strconv.FormatInt(*c.HostConfig.ShmSize, 10))
				}
			}
		}

		if c.HostConfig.Privileged {
			// Clear readonly for /sys.
			if mounts[i].Destination == "/sys" && !s.Root.Readonly {
				clearReadonly(&mounts[i])
			}

			// Clear readonly for cgroup
			if mounts[i].Type == "cgroup" {
				clearReadonly(&mounts[i])
			}
		}
	}

	s.Mounts = sortMounts(mounts)
	return nil
}

// generateNetworkMounts will generate network mounts.
func generateNetworkMounts(c *Container) []specs.Mount {
	mounts := make([]specs.Mount, 0)

	fileBinds := []struct {
		Name   string
		Source string
		Dest   string
	}{
		{"HostnamePath", c.HostnamePath, "/etc/hostname"},
		{"HostsPath", c.HostsPath, "/etc/hosts"},
		{"ResolvConfPath", c.ResolvConfPath, "/etc/resolv.conf"},
	}

	for _, bind := range fileBinds {
		if bind.Source != "" {
			_, err := os.Stat(bind.Source)
			if err != nil {
				logrus.Warnf("%s set to %s, but stat error: %v, skip it", bind.Name, bind.Source, err)
			} else {
				mounts = append(mounts, specs.Mount{
					Source:      bind.Source,
					Destination: bind.Dest,
					Type:        "bind",
					Options:     []string{"rbind", "rprivate"},
				})
			}
		}
	}

	return mounts
}

// trySetupNetworkMount will try to set network mount.
func trySetupNetworkMount(mount *types.MountPoint, c *Container) bool {
	if mount.Destination == "/etc/hostname" {
		c.HostnamePath = mount.Source
		return true
	}

	if mount.Destination == "/etc/hosts" {
		c.HostsPath = mount.Source
		return true
	}

	if mount.Destination == "/etc/resolv.conf" {
		c.ResolvConfPath = mount.Source
		return true
	}

	return false
}

// mounts defines how to sort specs.Mount.
type mounts []specs.Mount

// Len returns the number of mounts.
func (m mounts) Len() int {
	return len(m)
}

// Less returns true if the destination of mount i < destination of mount j
// in lexicographic order.
func (m mounts) Less(i, j int) bool {
	return filepath.Clean(m[i].Destination) < filepath.Clean(m[j].Destination)
}

// Swap swaps two items in an array of mounts.
func (m mounts) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// sortMounts sorts an array of mounts in lexicographic order. This ensure that
// the mount like /etc/resolv.conf will not mount before /etc, so /etc will
// not shadow /etc/resolv.conf
func sortMounts(m []specs.Mount) []specs.Mount {
	sort.Stable(mounts(m))
	return m
}
