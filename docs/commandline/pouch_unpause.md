## pouch unpause

Unpause a paused container

### Synopsis

Unpause a paused container in Pouchd. when unpausing, the paused container will resumes the process execution within the container.The container you unpaused will be running again if no error occurs.

```
pouch unpause CONTAINER
```

### Examples

```
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Paused   docker.io/library/busybox:latest   runc
$ pouch unpause foo
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running    docker.io/library/busybox:latest   runc
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

