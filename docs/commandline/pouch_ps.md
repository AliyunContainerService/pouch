## pouch ps

List all containers

### Synopsis


List Containers with container name, ID, status, image reference, runtime and creation time.

```
pouch ps [flags]
```

### Examples

```
$ pouch ps
Name     ID       Status    Image                              Runtime   Created
1dad17   1dad17   stopped   docker.io/library/busybox:latest   runv      about 1 hour ago
505571   505571   stopped   docker.io/library/busybox:latest   runc      about 1 hour ago

```

### Options

```
  -h, --help   help for ps
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

