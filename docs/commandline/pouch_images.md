## pouch images

List all images

### Synopsis

List all images in Pouchd.This is useful when you wish to have a look at images and Pouchd will show all local images with their NAME and SIZE.All local images will be shown in a table format you can use.

```
pouch images [flags]
```

### Examples

```
$ pouch images ls
IMAGE ID             IMAGE NAME                                               SIZE
bbc3a0323522         docker.io/library/busybox:latest                         703.14 KB
b81f317384d7         docker.io/library/nginx:latest                           42.39 MB
```

### Options

```
      --digest   Show images with digest
  -h, --help     help for images
  -q, --quiet    Only show image numeric ID
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

