## pouch cp

Copy files/folders between a container and the local filesystem

### Synopsis

Copy files/folders between a container and the local filesystem

Use '-' as the source to read a tar archive from stdin
and extract it to a directory destination in a container.
Use '-' as the destination to stream a tar archive of a
container source to stdout.

```
pouch cp [OPTIONS] CONTAINER:SRC_PATH DEST_PATH|-
  pouch cp [OPTIONS] SRC_PATH|- CONTAINER:DEST_PATH
```

### Examples

```
$ pouch cp 8assd1234:/root/foo /home
$ pouch cp /home/bar 712yasbc:/root
```

### Options

```
  -h, --help   help for cp
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

