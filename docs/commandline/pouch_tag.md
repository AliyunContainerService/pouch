## pouch tag

Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE

### Synopsis

tag command is to add tag reference for the existing image.

```
pouch tag SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]
```

### Examples

```
$ pouch tag registry.hub.docker.com/library/busybox:1.28 busybox:latest
```

### Options

```
  -h, --help   help for tag
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

