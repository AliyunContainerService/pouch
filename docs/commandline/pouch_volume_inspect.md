## pouch volume inspect

Inspect a pouch volume

### Synopsis

Inspect a volume in pouchd. It must specify volume's name.

```
pouch volume inspect [OPTIONS] NAME
```

### Examples

```
$ pouch volume inspect pouch-volume
Mountpoint:   /mnt/local/pouch-volume
Name:         pouch-volume
Scope:
CreatedAt:    2018-1-17 14:09:30
Driver:       local
```

### Options

```
  -h, --help   help for inspect
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

