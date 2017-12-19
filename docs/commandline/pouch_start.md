## pouch start

Start a created or stopped container

### Synopsis

Start a created container object in Pouchd. When starting, the relevant resource preserved during creating period comes into use.This is useful when you wish to start a container which has been created in advance.The container you started will be running if no error occurs.

```
pouch start [container] [flags]
```

### Examples

```
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Created   docker.io/library/busybox:latest   runc
$ pouch start foo
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running   docker.io/library/busybox:latest   runc
```

### Options

```
  -a, --attach        Attach container's STDOUT and STDERR
  -h, --help          help for start
  -i, --interactive   Attach container's STDIN
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

