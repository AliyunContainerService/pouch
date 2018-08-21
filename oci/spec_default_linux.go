package oci

import (
	"os"
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func iPtr(i int64) *int64        { return &i }
func u32Ptr(i int64) *uint32     { u := uint32(i); return &u }
func fmPtr(i int64) *os.FileMode { fm := os.FileMode(i); return &fm }

func newDefaultCaps() []string {
	return []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_FSETID",
		"CAP_KILL",
		"CAP_SETGID",
		"CAP_SETUID",
		"CAP_SETPCAP",
		"CAP_AUDIT_WRITE",
		"CAP_NET_BIND_SERVICE",
		"CAP_NET_RAW",
		"CAP_SYS_CHROOT",
		"CAP_MKNOD",
		"CAP_SETFCAP",
		"CAP_LINUX_IMMUTABLE", //add
		"CAP_NET_BROADCAST",   //add
		"CAP_NET_ADMIN",       //add
		"CAP_IPC_LOCK",        //add
		"CAP_IPC_OWNER",       //add
		"CAP_SYS_MODULE",      //add
		"CAP_SYS_PTRACE",      //add
		"CAP_SYS_PACCT",       //add
		"CAP_SYS_ADMIN",       //add
		"CAP_SYS_NICE",        //add
		"CAP_SYS_RESOURCE",    //add
		"CAP_SYS_TTY_CONFIG",  //add
		"CAP_LEASE",           //add
		"CAP_AUDIT_CONTROL",   //add
		"CAP_SYSLOG",          //add
		"CAP_WAKE_ALARM",      //add
		"CAP_BLOCK_SUSPEND",   //add
		"CAP_DAC_READ_SEARCH", //add
	}
}

// NewDefaultSpec create a new spec.
func NewDefaultSpec() *specs.Spec {
	s := &specs.Spec{
		Version: specs.Version,
	}
	s.Mounts = []specs.Mount{
		{
			Destination: "/proc",
			Type:        "proc",
			Source:      "proc",
			Options:     []string{"nosuid", "noexec", "nodev"},
		},
		{
			Destination: "/dev",
			Type:        "tmpfs",
			Source:      "tmpfs",
			Options:     []string{"nosuid", "strictatime", "mode=755"},
		},
		{
			Destination: "/dev/pts",
			Type:        "devpts",
			Source:      "devpts",
			Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
		},
		{
			Destination: "/sys",
			Type:        "sysfs",
			Source:      "sysfs",
			Options:     []string{"nosuid", "noexec", "nodev", "ro"},
		},
		{
			Destination: "/dev/mqueue",
			Type:        "mqueue",
			Source:      "mqueue",
			Options:     []string{"nosuid", "noexec", "nodev"},
		},
		// TODO(huamin): shm need to remove in default spec
		{
			Destination: "/dev/shm",
			Type:        "tmpfs",
			Source:      "shm",
			Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
		},
	}

	if _, err := os.Stat("/proc/self/ns/net"); err == nil {
		s.Mounts = append(s.Mounts, specs.Mount{
			Destination: "/sys/fs/cgroup",
			Type:        "cgroup",
			Source:      "cgroup",
			Options:     []string{"ro", "nosuid", "noexec", "nodev"},
		})
	}

	if defaultCaps := os.Getenv("default_caps"); defaultCaps != "" {
		list := strings.Split(defaultCaps, ":")
		s.Process = &specs.Process{
			Capabilities: &specs.LinuxCapabilities{
				Bounding:    list,
				Permitted:   list,
				Inheritable: list,
				Effective:   list,
			},
		}
	} else {
		s.Process = &specs.Process{
			Capabilities: &specs.LinuxCapabilities{
				Bounding:    newDefaultCaps(),
				Permitted:   newDefaultCaps(),
				Inheritable: newDefaultCaps(),
				Effective:   newDefaultCaps(),
			},
		}
	}

	s.Linux = &specs.Linux{
		MaskedPaths: []string{
			"/proc/kcore",
			"/proc/latency_stats",
			"/proc/timer_list",
			"/proc/timer_stats",
			"/proc/sched_debug",
		},
		ReadonlyPaths: []string{
			"/proc/asound",
			"/proc/bus",
			"/proc/fs",
			"/proc/irq",
			"/proc/sys",
			"/proc/sysrq-trigger",
		},
		Namespaces: []specs.LinuxNamespace{
			{
				Type: specs.PIDNamespace,
			},
			{
				Type: specs.IPCNamespace,
			},
			{
				Type: specs.UTSNamespace,
			},
			{
				Type: specs.MountNamespace,
			},
			{
				Type: specs.NetworkNamespace,
			},
		},
		// Devices implicitly contains the following devices:
		// null, zero, full, random, urandom, tty, console, and ptmx.
		// ptmx is a bind-mount or symlink of the container's ptmx.
		// See also: https://github.com/opencontainers/runtime-spec/blob/master/config-linux.md#default-devices
		Devices: []specs.LinuxDevice{
			{
				Type:     "c",
				Path:     "/dev/fuse",
				Major:    10,
				Minor:    229,
				FileMode: fmPtr(0666),
				UID:      u32Ptr(0),
				GID:      u32Ptr(0),
			},
		},
		Resources: &specs.LinuxResources{
			Devices: []specs.LinuxDeviceCgroup{
				{
					Allow:  false,
					Access: "rwm",
				},
				{
					Allow:  true,
					Type:   "c",
					Major:  iPtr(1),
					Minor:  iPtr(5),
					Access: "rwm",
				},
				{
					Allow:  true,
					Type:   "c",
					Major:  iPtr(1),
					Minor:  iPtr(3),
					Access: "rwm",
				},
				{
					Allow:  true,
					Type:   "c",
					Major:  iPtr(1),
					Minor:  iPtr(9),
					Access: "rwm",
				},
				{
					Allow:  true,
					Type:   "c",
					Major:  iPtr(1),
					Minor:  iPtr(8),
					Access: "rwm",
				},
				{
					Allow:  true,
					Type:   "c",
					Major:  iPtr(5),
					Minor:  iPtr(0),
					Access: "rwm",
				},

				{
					Allow:  true,
					Type:   "c",
					Major:  iPtr(5),
					Minor:  iPtr(1),
					Access: "rwm",
				},
				{
					Allow:  false,
					Type:   "c",
					Major:  iPtr(10),
					Minor:  iPtr(229),
					Access: "rwm",
				},
			},
		},
	}

	return s

}
