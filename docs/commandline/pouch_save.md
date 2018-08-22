## pouch save

Save an image to a tar archive or STDOUT

### Synopsis

save an image to a tar archive.

```
pouch save [OPTIONS] IMAGE
```

### Examples

```
$ pouch save -o busybox.tar busybox:latest
$ pouch load -i busybox.tar foo
$ pouch images
IMAGE ID       IMAGE NAME                                           SIZE
8c811b4aec35   registry.hub.docker.com/library/busybox:latest       710.81 KB
8c811b4aec35   foo:latest                                           710.81 KB

```

### Options

```
  -h, --help            help for save
  -o, --output string   Save to a tar archive file, instead of STDOUT
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

