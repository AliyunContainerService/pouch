## pouch volume remove

Remove a volume

### Synopsis

Remove a volume in pouchd. It must specify volume's name, and the volume will be removed when it is existent and unused.

```
pouch volume remove [OPTIONS] NAME
```

### Examples

```
$ pouch volume rm pouch-volume
Removed: pouch-volume
```

### Options

```
  -h, --help   help for remove
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

* [pouch volume](pouch_volume.md)	 - Manage pouch volumes

