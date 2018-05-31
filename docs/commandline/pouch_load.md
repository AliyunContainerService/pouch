## pouch load

load a set of images from a tar archive or STDIN

### Synopsis

load a set of images by tar stream

```
pouch load [OPTIONS] [IMAGE_NAME]
```

### Examples

```
$ pouch load -i busybox.tar busybox
```

### Options

```
  -h, --help           help for load
  -i, --input string   Read from tar archive file, instead of STDIN
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

