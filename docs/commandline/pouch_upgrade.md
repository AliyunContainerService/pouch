## pouch upgrade

Upgrade a container with new image and args

### Synopsis

upgrade is a feature to replace a container's image. You can specify the new Entrypoint and Cmd for the new container. When you want to update a container's image, but inherit the network and volumes of the old container, then you should think about the upgrade feature.

```
pouch upgrade [OPTIONS] CONTAINER [COMMAND] [ARG...]
```

### Examples

```
 $ pouch run -d -m 20m --name test  registry.hub.docker.com/library/busybox:latest
4c58d27f58d38776dda31c01c897bbf554c802a9b80ae4dc20be1337f8a969f2
$ pouch upgrade --image registry.hub.docker.com/library/hello-world:latest test
test
```

### Options

```
      --entrypoint string   Overwrite the default ENTRYPOINT of the image
  -h, --help                help for upgrade
      --image string        Specify image of the new container
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

