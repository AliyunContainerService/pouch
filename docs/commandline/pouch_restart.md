## pouch restart

restart one or more containers

### Synopsis

restart one or more containers

```
pouch restart [OPTION] CONTAINER [CONTAINER...]
```

### Examples

```
$ pouch ps -a
Name     ID       Status    Image                              Runtime
foo      71b9c1   Stopped   docker.io/library/busybox:latest   runc
$ pouch restart foo
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running   docker.io/library/busybox:latest   runc
```

### Options

```
  -h, --help       help for restart
  -t, --time int   Seconds to wait for stop before killing the container (default 10)
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

