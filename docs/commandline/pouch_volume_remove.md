## pouch volume remove

Remove volume

### Synopsis

Remove a volume in pouchd. It need specify volume's name, when the volume is exist and is unuse, it will be remove.

```
pouch volume remove <name> [flags]
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
  -H, --host string        Specify connecting address of Pouch CLI (default "unix:///var/run/pouchd.sock")
      --tlscacert string   Specify CA file of TLS
      --tlscert string     Specify cert file of TLS
      --tlskey string      Specify key file of TLS
      --tlsverify          Use TLS and verify remote
```

### SEE ALSO

* [pouch volume](pouch_volume.md)	 - Manage pouch volumes

