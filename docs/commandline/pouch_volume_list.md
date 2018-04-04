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
DRIVER   VOLUME NAME
local    pouch-volume-1
local    pouch-volume-2
local    pouch-volume-3
```

### Options

```
  -h, --help         help for list
      --mountpoint   Display volume mountpoint
      --size         Display volume size
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

