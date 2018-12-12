package mgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/alibaba/pouch/apis/types"

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

// setupMounts create mount spec.
func setupMounts(ctx context.Context, c *Container, s *specs.Spec) error {
	var mounts []specs.Mount
	// Override the default mounts which are duplicate with user defined ones.
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
		if sm.Destination == "/dev/shm" && c.HostConfig.ShmSize != nil {
			sm.Options = append(sm.Options, fmt.Sprintf("size=%s", strconv.FormatInt(*c.HostConfig.ShmSize, 10)))
		}
		mounts = append(mounts, sm)
	}
	// TODO: we can suggest containerd to add the cgroup into the default spec.
	mounts = append(mounts, specs.Mount{
		Destination: "/sys/fs/cgroup",
		Type:        "cgroup",
		Source:      "cgroup",
		Options:     []string{"ro", "nosuid", "noexec", "nodev"},
	})

	if c.HostConfig == nil {
		return nil
	}
	// user defined mount
	for _, mp := range c.Mounts {
		if trySetupNetworkMount(mp, c) {
			// ignore the network mount, we will handle it later.
			continue
		}

		// check duplicate mountpoint
		for _, sm := range mounts {
			if sm.Destination == mp.Destination {
				return fmt.Errorf("duplicate mount point: %s", mp.Destination)
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
		if pg != "" {
			opts = append(opts, pg)
		}

		// TODO: support copy data.

		if mp.Destination == "/dev/shm" && c.HostConfig.ShmSize != nil {
			opts = []string{fmt.Sprintf("size=%s", strconv.FormatInt(*c.HostConfig.ShmSize, 10))}
		}

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

	s.Mounts = sortMounts(mounts)

	if c.HostConfig.Privileged {
		for i := range s.Mounts {
			// Clear readonly for /sys.
			if s.Mounts[i].Destination == "/sys" && !s.Root.Readonly {
				clearReadonly(&s.Mounts[i])
			}

			// Clear readonly for cgroup
			if s.Mounts[i].Type == "cgroup" {
				clearReadonly(&s.Mounts[i])
			}
		}
	}
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
