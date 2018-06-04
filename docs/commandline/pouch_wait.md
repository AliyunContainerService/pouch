## pouch wait

Block until one or more containers stop, then print their exit codes

### Synopsis

Block until one or more containers stop, then print their exit codes. If container state is already stopped, the command will return exit code immediately. On a successful stop, the exit code of the container is returned. 

```
pouch wait CONTAINER [CONTAINER...]
```

### Examples

```
$ pouch ps
Name   ID       Status         Created         Image                                            Runtime
foo    f6717e   Up 2 seconds   3 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch stop foo
$ pouch ps -a
Name   ID       Status                 Created         Image                                            Runtime
foo    f6717e   Stopped (0) 1 minute   2 minutes ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch wait foo
0
```

### Options

```
  -h, --help   help for wait
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

