## pouch create

Create a new container with specified image

### Synopsis

Create a static container object in Pouchd. When creating, all configuration user input will be stored in memory store of Pouchd. This is useful when you wish to create a container configuration ahead of time so that Pouchd will preserve the resource in advance. The container you created is ready to start when you need it.

```
pouch create [OPTIONS] IMAGE [ARG...]
```

### Examples

```
$ pouch create --name foo busybox:latest
container ID: e1d541722d68dc5d133cca9e7bd8fd9338603e1763096c8e853522b60d11f7b9, name: foo
```

### Options

```
      --cpu-share int          CPU shares
      --cpuset-cpus string     CPUs in cpuset
      --cpuset-mems string     MEMs in cpuset
      --device strings         Add a host device to the container
      --enableLxcfs            Enable lxcfs
      --entrypoint string      Overwrite the default entrypoint
  -e, --env strings            Set environment variables for container
  -h, --help                   help for create
      --hostname string        Set container's hostname
  -l, --label strings          Set label for a container
  -m, --memory string          Container memory limit
      --memory-swap string     Container swap limit
      --memory-wappiness int   Container memory swappiness [0, 100] (default -1)
      --name string            Specify name of container
      --restart string         Restart policy to apply when container exits
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

