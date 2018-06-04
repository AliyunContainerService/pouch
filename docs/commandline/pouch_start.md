## pouch start

Start one or more created or stopped containers

### Synopsis

Start one or more created container objects in Pouchd. When starting, the relevant resource preserved during creating period comes into use.This is useful when you wish to start a container which has been created in advance.The container you started will be running if no error occurs.

```
pouch start [OPTIONS] CONTAINER [CONTAINER...]
```

### Examples

```
$ pouch ps -a
Name   ID       Status    Created         Image                                            Runtime
foo2   5a0ede   created   1 second ago    registry.hub.docker.com/library/busybox:latest   runc
foo1   e05637   created   6 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch start foo1 foo2
foo1
foo2
$ pouch ps
Name   ID       Status         Created          Image                                            Runtime
foo2   5a0ede   Up 2 seconds   12 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
foo1   e05637   Up 3 seconds   17 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
```

### Options

```
  -a, --attach               Attach container's STDOUT and STDERR
      --detach-keys string   Override the key sequence for detaching a container
  -h, --help                 help for start
  -i, --interactive          Attach container's STDIN
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

