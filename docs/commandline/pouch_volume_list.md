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
$ pouch volume list --quiet
VOLUME NAME
pouch-volume-1
pouch-volume-2
pouch-volume-3
```

### Options

```
  -f, --filter strings   Filter output based on conditions provided, filter support driver, name, label
  -h, --help             help for list
      --mountpoint       Display volume mountpoint
  -q, --quiet            Only display volume names
      --size             Display volume size
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

