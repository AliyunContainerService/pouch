## pouch pull

Pull an image from registry

### Synopsis

Pull an image or a repository from a registry. Most of your images will be created on top of a base image from the registry. So, you can pull and try prebuilt images contained by registry without needing to define and configure your own.

```
pouch pull IMAGE
```

### Examples

```
$ pouch images
IMAGE ID            IMAGE NAME                           SIZE
bbc3a0323522        docker.io/library/busybox:latest     703.14 KB
$ pouch pull docker.io/library/redis:alpine
$ pouch images
IMAGE ID            IMAGE NAME                           SIZE
bbc3a0323522        docker.io/library/busybox:latest     703.14 KB
0153c5db97e5        docker.io/library/redis:alpine       9.63 MB
```

### Options

```
  -h, --help   help for pull
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

