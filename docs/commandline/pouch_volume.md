## pouch volume

Manage pouch volumes

### Synopsis

Manager the volumes in pouchd. It contains the functions of create/remove/list/inspect volume, 'driver' is used to list drivers that pouch support. The default volume driver is local, it will make a directory to bind into container.

```
pouch volume [command]
```

### Options

```
  -h, --help   help for volume
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
* [pouch volume create](pouch_volume_create.md)	 - Create a volume
* [pouch volume inspect](pouch_volume_inspect.md)	 - Inspect one or more pouch volumes
* [pouch volume list](pouch_volume_list.md)	 - List volumes
* [pouch volume remove](pouch_volume_remove.md)	 - Remove a volume

