## pouch kill

Kill one or more running containers

### Synopsis

Kill one or more running containers, the container will receive SIGKILL by default, or the signal which is specified with the --signal option.

```
pouch kill [OPTIONS] CONTAINER [CONTAINER...]
```

### Examples

```
$ pouch ps
Name            ID       Status          Created          Image                                            Runtime
foo             c926cf   Up 5 seconds    6 seconds ago    registry.hub.docker.com/library/busybox:latest   runc
$ pouch kill foo
foo
$ pouch ps -a
Name            ID       Status                     Created          Image                                            Runtime
foo             c926cf   Exited (137) 9 seconds     25 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
```

### Options

```
  -h, --help            help for kill
  -s, --signal string   Signal to send to the container (default "SIGKILL")
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

