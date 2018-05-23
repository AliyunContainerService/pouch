## pouch unpause

Unpause one or more paused container

### Synopsis

Unpause one or more paused containers in Pouchd. when unpausing, the paused container will resumes the process execution within the container.The container you unpaused will be running again if no error occurs.

```
pouch unpause CONTAINER [CONTAINER...]
```

### Examples

```
$ pouch ps
Name   ID       Status                  Created          Image                                            Runtime
foo2   c95673   Up 13 seconds(paused)   14 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
foo1   204cc6   Up 17 seconds(paused)   17 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch unpause foo1 foo2
foo1
foo2
$ pouch ps
Name   ID       Status          Created          Image                                            Runtime
foo2   c95673   Up 48 seconds   49 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
foo1   204cc6   Up 52 seconds   52 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
```

### Options

```
  -h, --help   help for unpause
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

