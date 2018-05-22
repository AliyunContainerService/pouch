## pouch rm

Remove one or more containers

### Synopsis


Remove one or more containers in Pouchd.
If a container be stopped or created, you can remove it. 
If the container is running, you can also remove it with flag force.
When the container is removed, the all resources of the container will
be released.


```
pouch rm [OPTIONS] CONTAINER [CONTAINER...]
```

### Examples

```
$ pouch ps -a
Name   ID       Status                  Created          Image                                            Runtime
foo    03cd58   Exited (0) 25 seconds   26 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch rm foo
foo
$ pouch ps
Name   ID       Status         Created          Image                                            Runtime
foo2   1d979d   Up 5 seconds   6 seconds ago    registry.hub.docker.com/library/busybox:latest   runc
foo1   83e3cf   Up 9 seconds   10 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch rm -f foo1 foo2
foo1
foo2
```

### Options

```
  -f, --force     if the container is running, force to remove it
  -h, --help      help for rm
  -v, --volumes   remove container's volumes that create by the container
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

