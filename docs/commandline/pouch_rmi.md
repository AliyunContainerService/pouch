## pouch rmi

Remove one or more images by reference

### Synopsis

Remove one or more images by reference.When the image is being used by a container, you must specify -f to delete it.But it is strongly discouraged, because the container will be in abnormal status.

```
pouch rmi [OPTIONS] IMAGE [IMAGE...]
```

### Examples

```
$ pouch rmi registry.hub.docker.com/library/busybox:latest
registry.hub.docker.com/library/busybox:latest
$ pouch create --name test registry.hub.docker.com/library/busybox:latest
container ID: e5952417f9ee94621bbeaec532be1803ae2dedeb11a80f578a6d621e04a95afd, name: test
$ pouch rmi registry.hub.docker.com/library/busybox:latest
Error: failed to remove image: {"message":"Unable to remove the image \"registry.hub.docker.com/library/busybox:latest\" (must force) - container e5952417f9ee94621bbeaec532be1803ae2dedeb11a80f578a6d621e04a95afd is using this image"}

```

### Options

```
  -f, --force   if image is being used, remove image and all associated resources
  -h, --help    help for rmi
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

