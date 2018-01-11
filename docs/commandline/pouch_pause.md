## pouch pause

Pause a running container

### Synopsis

Pause a running container object in Pouchd. when pausing, the container will pause its running but hold all the relevant resource.This is useful when you wish to pause a container for a while and to restore the running status later.The container you paused will pause without being terminated.

```
pouch pause CONTAINER
```

### Examples

```
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running   docker.io/library/busybox:latest   runc
$ pouch pause foo
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Paused    docker.io/library/busybox:latest   runc
```

### Options

```
  -h, --help   help for pause
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

