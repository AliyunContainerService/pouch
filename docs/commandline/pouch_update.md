## pouch update

Update the configurations of a container

### Synopsis

Update a container's configurations, including memory, cpu and diskquota etc.  You can update a container when it is running.

```
pouch update [OPTIONS] CONTAINER
```

### Examples

```
$ pouch run -d -m 20m --name test-update registry.hub.docker.com/library/busybox:latest
8649804cb63ff9713a2734d99728b9d6d5d1e4d2fbafb2b4dbdf79c6bbaef812
$ cat /sys/fs/cgroup/memory/8649804cb63ff9713a2734d99728b9d6d5d1e4d2fbafb2b4dbdf79c6bbaef812/memory.limit_in_bytes
20971520
$ pouch update -m 30m test-update
$ cat /sys/fs/cgroup/memory/8649804cb63ff9713a2734d99728b9d6d5d1e4d2fbafb2b4dbdf79c6bbaef812/memory.limit_in_bytes
31457280
	
```

### Options

```
      --blkio-weight uint16     Block IO (relative weight), between 10 and 1000, or 0 to disable
      --cpu-period int          Limit CPU CFS (Completely Fair Scheduler) period, range is in [1000(1ms),1000000(1s)]
      --cpu-share int           CPU shares (relative weight)
      --cpuset-cpus string      CPUs in cpuset
      --cpuset-mems string      MEMs in cpuset
  -e, --env strings             Set environment variables for container
  -h, --help                    help for update
  -l, --label strings           Set label for container
  -m, --memory string           Container memory limit
      --memory-swap string      Container swap limit
      --memory-swappiness int   Container memory swappiness [0, 100] (default -1)
      --quota string            Update disk quota for container
      --restart string          Restart policy to apply when container exits
```

### Options inherited from parent commands

```
  -H, --host string        Specify connecting address of Pouch CLI (default "unix:///var/run/pouchd.sock")
      --tlscacert string   Specify CA file of TLS
      --tlscert string     Specify cert file of TLS
      --tlskey string      Specify key file of TLS
      --tlsverify          Use TLS and verify remote
```

### SEE ALSO

* [pouch](pouch.md)	 - An efficient container engine

