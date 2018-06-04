## pouch remount-lxcfs

remount lxcfs bind in containers

### Synopsis


remount lxcfs in containers.

```
pouch remount-lxcfs
```

### Examples

```
$ pouch remount-lxcfs
ID       Status
e42c68   OK

```

### Options

```
  -h, --help   help for remount-lxcfs
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

