## pouch pause

Pause one or more running containers

### Synopsis

Pause one or more running containers in Pouchd. when pausing, the container will pause its running but hold all the relevant resource. This is useful when you wish to pause a container for a while and to restore the running status later. The container you paused will pause without being terminated.

```
pouch pause CONTAINER [CONTAINER...]
```

### Examples

```
$ pouch ps
Name   ID       Status          Created          Image                                            Runtime
foo2   87259c   Up 25 seconds   26 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
foo1   77188c   Up 46 seconds   47 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch pause foo1 foo2
foo1
foo2
$ pouch ps
Name   ID       Status                Created        Image                                            Runtime
foo2   87259c   Up 1 minute(paused)   1 minute ago   registry.hub.docker.com/library/busybox:latest   runc
foo1   77188c   Up 1 minute(paused)   1 minute ago   registry.hub.docker.com/library/busybox:latest   runc
```

### Options

```
  -h, --help   help for pause
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

