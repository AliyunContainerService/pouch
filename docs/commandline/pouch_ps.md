## pouch ps

List containers

### Synopsis


List Containers with container name, ID, status, creation time, image reference and runtime.

```
pouch ps [OPTIONS]
```

### Examples

```
$ pouch ps
Name   ID       Status          Created          Image                              Runtime
2      e42c68   Up 15 minutes   16 minutes ago   docker.io/library/busybox:latest   runc
1      a8c2ea   Up 16 minutes   17 minutes ago   docker.io/library/busybox:latest   runc

$ pouch ps -a
Name   ID       Status          Created          Image                              Runtime
3      faf132   created         16 seconds ago   docker.io/library/busybox:latest   runc
2      e42c68   Up 16 minutes   16 minutes ago   docker.io/library/busybox:latest   runc
1      a8c2ea   Up 17 minutes   18 minutes ago   docker.io/library/busybox:latest   runc

$ pouch ps -q
e42c68
a8c2ea

$ pouch ps -a -q
faf132
e42c68
a8c2ea

```

### Options

```
  -a, --all     Show all containers (default shows just running)
  -h, --help    help for ps
  -q, --quiet   Only show numeric IDs
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

