## pouch run

Create a new container and start it

### Synopsis

Create a container object in Pouchd, and start the container. This is useful when you just want to use one command to start a container. 

```
pouch run a container [flags]
```

### Examples

```
$ pouch run --name test registry.hub.docker.com/library/busybox:latest echo "hi"
23f8529fddf7c8bbea70e2c12353e47dbfa5eacda9d58ff8665269614456424b
$ pouch ps
Name   ID       Status    Image                                            Runtime   Created
test   23f852   stopped   registry.hub.docker.com/library/busybox:latest   runc      4 seconds ago
$ pouch run -i --name test registry.hub.docker.com/library/busybox:latest echo "hi"
hi
$ pouch ps
Name   ID       Status    Image                                            Runtime   Created
test   883ea9   stopped   registry.hub.docker.com/library/busybox:latest   runc      5 seconds ago
	
```

### Options

```
  -a, --attach               Attach container's STDOUT and STDERR
      --detach-keys string   Override the key sequence for detaching a container
  -h, --help                 help for run
  -i, --interactive          Attach container's STDIN
      --name string          Specify name of container
      --runtime string       Specify oci runtime
  -t, --tty                  Allocate a tty device
  -v, --volume strings       Bind mount volumes to container
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

