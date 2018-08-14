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

$ pouch ps --no-trunc
Name   ID                                                                 Status        Created        Image                            Runtime
foo2   692c77587b38f60bbd91d986ec3703848d72aea5030e320d4988eb02aa3f9d48   Up 1 minute   1 minute ago   docker.io/library/redis:alpine   runc
foo    18592900006405ee64788bd108ef1de3d24dc3add73725891f4787d0f8e036f5   Up 1 minute   1 minute ago   docker.io/library/redis:alpine   runc

$ pouch ps --no-trunc -q
692c77587b38f60bbd91d986ec3703848d72aea5030e320d4988eb02aa3f9d48
18592900006405ee64788bd108ef1de3d24dc3add73725891f4787d0f8e036f5

$ pouch ps --no-trunc -a
Name   ID                                                                 Status         Created         Image                            Runtime
foo3   63fd6371f3d614bb1ecad2780972d5975ca1ab534ec280c5f7d8f4c7b2e9989d   created        2 minutes ago   docker.io/library/redis:alpine   runc
foo2   692c77587b38f60bbd91d986ec3703848d72aea5030e320d4988eb02aa3f9d48   Up 2 minutes   2 minutes ago   docker.io/library/redis:alpine   runc
foo    18592900006405ee64788bd108ef1de3d24dc3add73725891f4787d0f8e036f5   Up 2 minutes   2 minutes ago   docker.io/library/redis:alpine   runc

```

### Options

```
  -a, --all              Show all containers (default shows just running)
  -f, --filter strings   Filter output based on given conditions, support filter key [ id label name status ]
  -h, --help             help for ps
      --no-trunc         Do not truncate output
  -q, --quiet            Only show numeric IDs
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

