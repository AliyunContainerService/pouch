## pouch run

Create a new container and start it

### Synopsis

Create a container object in Pouchd, and start the container. This is useful when you just want to use one command to start a container. 

```
pouch run [OPTIONS] IMAGE [ARG...]
```

### Examples

```
$ pouch run --name test registry.hub.docker.com/library/busybox:latest echo "hi"
hi
$ pouch ps
Name   ID       Status    Image                                            Runtime   Created
test   23f852   stopped   registry.hub.docker.com/library/busybox:latest   runc      4 seconds ago
$ pouch run -d --name test registry.hub.docker.com/library/busybox:latest
90719b5f9a455b3314a49e72e3ecb9962f215e0f90153aa8911882acf2ba2c84
$ pouch ps
Name   ID       Status    Image                                            Runtime   Created
test   90719b   stopped   registry.hub.docker.com/library/busybox:latest   runc      5 seconds ago
$ pouch run --device /dev/zero:/dev/testDev:rwm --name test registry.hub.docker.com/library/busybox:latest ls -l /dev/testDev
crw-rw-rw-    1 root     root        1,   3 Jan  8 09:40 /dev/testnull
	
```

### Options

```
  -a, --attach                 Attach container's STDOUT and STDERR
      --cpu-share int          CPU shares
      --cpuset-cpus string     CPUs in cpuset
      --cpuset-mems string     MEMs in cpuset
  -d, --detach                 Run container in background and print container ID
      --detach-keys string     Override the key sequence for detaching a container
      --device strings         Add a host device to the container
      --enableLxcfs            Enable lxcfs
      --entrypoint string      Overwrite the default entrypoint
  -e, --env strings            Set environment variables for container
  -h, --help                   help for run
      --hostname string        Set container's hostname
  -i, --interactive            Attach container's STDIN
  -l, --label strings          Set label for a container
  -m, --memory string          Container memory limit
      --memory-swap string     Container swap limit
      --memory-wappiness int   Container memory swappiness [0, 100] (default -1)
      --name string            Specify name of container
      --runtime string         Specify oci runtime
  -t, --tty                    Allocate a tty device
  -v, --volume strings         Bind mount volumes to container
  -w, --workdir string         Set the working directory in a container
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

