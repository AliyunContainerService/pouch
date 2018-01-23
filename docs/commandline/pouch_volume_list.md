## pouch volume list

List volumes

### Synopsis

List volumes in pouchd. It lists the volume's name

```
pouch volume list
```

### Examples

```
$ pouch volume list
Name:
pouch-volume-1
pouch-volume-2
pouch-volume-3
```

### Options

```
  -h, --help   help for list
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

