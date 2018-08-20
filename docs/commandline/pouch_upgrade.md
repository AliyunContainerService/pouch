## pouch upgrade

Upgrade a container with new image and args

### Synopsis

Upgrade a container with new image and args

```
pouch upgrade [OPTIONS] IMAGE [COMMAND] [ARG...]
```

### Examples

```
 $ pouch run -d -m 20m --name test1  registry.hub.docker.com/library/busybox:latest
4c58d27f58d38776dda31c01c897bbf554c802a9b80ae4dc20be1337f8a969f2
$ pouch upgrade --name test1 registry.hub.docker.com/library/hello-world:latest
test1
```

### Options

```
      --annotation strings           Additional annotation for runtime
      --blkio-weight uint16          Block IO (relative weight), between 10 and 1000, or 0 to disable
      --blkio-weight-device value    Block IO weight (relative device weight), need CFQ IO Scheduler enable (default [])
      --cap-add strings              Add Linux capabilities
      --cap-drop strings             Drop Linux capabilities
      --cgroup-parent string         Optional parent cgroup for the container
      --cpu-period int               Limit CPU CFS (Completely Fair Scheduler) period, range is in [1000(1ms),1000000(1s)]
      --cpu-quota int                Limit CPU CFS (Completely Fair Scheduler) quota, range is in [1000,∞)
      --cpu-shares int               CPU shares (relative weight)
      --cpuset-cpus string           CPUs in which to allow execution (0-3, 0,1)
      --cpuset-mems string           MEMs in which to allow execution (0-3, 0,1)
      --device strings               Add a host device to the container
      --device-read-bps value        Limit read rate (bytes per second) from a device (default [])
      --device-read-iops value       Limit read rate (IO per second) from a device (default [])
      --device-write-bps value       Limit write rate (bytes per second) from a device (default [])
      --device-write-iops value      Limit write rate (IO per second) from a device (default [])
      --disable-network-files        Disable the generation of network files(/etc/hostname, /etc/hosts and /etc/resolv.conf) for container. If true, no network files will be generated. Default false
      --disk-quota strings           Set disk quota for container
      --enableLxcfs                  Enable lxcfs for the container, only effective when enable-lxcfs switched on in Pouchd
      --entrypoint string            Overwrite the default ENTRYPOINT of the image
  -e, --env strings                  Set environment variables for container
      --expose strings               Set expose container's ports
      --group-add strings            Add additional groups to join
  -h, --help                         help for upgrade
      --hostname string              Set container's hostname
      --initscript string            Initial script executed in container
      --intel-rdt-l3-cbm string      Limit container resource for Intel RDT/CAT which introduced in Linux 4.10 kernel
      --ipc string                   IPC namespace to use
  -l, --label strings                Set labels for a container
      --log-driver string            Logging driver for the container (default "json-file")
      --log-opt stringArray          Log driver options
  -m, --memory string                Memory limit
      --memory-extra int             Represent container's memory high water mark percentage, range in [0, 100]
      --memory-force-empty-ctl int   Whether to reclaim page cache when deleting the cgroup of container
      --memory-swap string           Swap limit equal to memory + swap, '-1' to enable unlimited swap
      --memory-swappiness int        Container memory swappiness [0, 100]
      --memory-wmark-ratio int       Represent this container's memory low water mark percentage, range in [0, 100]. The value of memory low water mark is memory.limit_in_bytes * MemoryWmarkRatio
      --name string                  Specify name of container
      --net strings                  Set networks to container
      --net-priority int             net priority
      --nvidia-capabilities string   NvidiaDriverCapabilities controls which driver libraries/binaries will be mounted inside the container
      --nvidia-visible-devs string   NvidiaVisibleDevices controls which GPUs will be made accessible inside the container
      --oom-kill-disable             Disable OOM Killer
      --oom-score-adj int            Tune host's OOM preferences (-1000 to 1000) (default -500)
      --pid string                   PID namespace to use
      --pids-limit int               Set container pids limit
  -p, --port strings                 Set container ports mapping
      --privileged                   Give extended privileges to the container
      --quota-id string              Specified quota id, if id < 0, it means pouchd alloc a unique quota id
      --restart string               Restart policy to apply when container exits
      --rich                         Start container in rich container mode. (default false)
      --rich-mode string             Choose one rich container mode. dumb-init(default), systemd, sbin-init
      --runtime string               OCI runtime to use for this container
      --sche-lat-switch int          Whether to enable scheduler latency count in cpuacct
      --security-opt strings         Security Options
      --shm-size string              Size of /dev/shm, default value is 64MB
      --sysctl strings               Sysctl options
  -t, --tty                          Allocate a pseudo-TTY
      --ulimit value                 Set container ulimit (default [])
  -u, --user string                  UID
      --uts string                   UTS namespace to use
  -v, --volume strings               Bind mount volumes to container, format is: [source:]<destination>[:mode], [source] can be volume or host's path, <destination> is container's path, [mode] can be "ro/rw/dr/rr/z/Z/nocopy/private/rprivate/slave/rslave/shared/rshared"
      --volumes-from strings         set volumes from other containers, format is <container>[:mode]
  -w, --workdir string               Set the working directory in a container
```

### Options inherited from parent commands

```
  -D, --debug              Switch client log level to DEBUG mode
  -H, --host string        Specify connecting address of Pouch CLI (default "unix:///var/run/pouchd.sock")
      --tlscacert string   Specify CA file of TLS
      --tlscert string     Specify cert file of TLS
      --tlskey string      Specify key file of TLS
      --tlsverify          Use TLS and verify remote
```

### SEE ALSO

* [pouch](pouch.md)	 - An efficient container engine

