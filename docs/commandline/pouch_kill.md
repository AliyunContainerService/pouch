## pouch kill

kill one or more running containers

### Synopsis

Kill one or more running container objects in Pouchd. You can kill a container using the container’s ID, ID-prefix, or name. This is useful when you wish to kill a container which is running.

```
pouch kill [OPTIONS] CONTAINER [CONTAINER...]
```

### Examples

```
$ pouch ps -a
Name   ID       Status    		Created         Image                                            Runtime
foo2   5a0ede   Up 2 seconds   	3 second ago    registry.hub.docker.com/library/busybox:latest   runc
foo1   e05637   Up 6 seconds   	7 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch kill foo1
foo1
$ pouch ps
Name   ID       Status         Created          Image                                            Runtime
foo2   5a0ede   Up 11 seconds   12 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
```

### Options

```
  -h, --help            help for kill
  -s, --signal string   Signal to send to the container (default "KILL") (default "KILL")
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

