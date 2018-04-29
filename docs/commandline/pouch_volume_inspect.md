## pouch volume inspect

Inspect one or more pouch volumes

### Synopsis

Inspect one or more volumes in pouchd. It must specify volume's name.

```
pouch volume inspect [OPTIONS] Volume [Volume...]
```

### Examples

```
$ pouch volume inspect pouch-volume
{
    "CreatedAt": "2018-4-2 14:33:45",
    "Driver": "local",
    "Labels": {
        "backend": "local",
        "hostname": "ubuntu"
    },
    "Mountpoint": "/mnt/local/pouch-volume",
    "Name": "pouch-volume",
    "Status": {
        "sifter": "Default",
        "size": "10g"
    }
}
```

### Options

```
  -f, --format string   Format the output using the given go template
  -h, --help            help for inspect
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

