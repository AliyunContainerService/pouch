## pouch network remove

Remove a pouch network

### Synopsis

Remove a network in pouchd. It must specify network's name.

```
pouch network remove [OPTIONS] NAME
```

### Examples

```
$ pouch network remove pouch-net
Removed: pouch-net
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

* [pouch network](pouch_network.md)	 - Manage pouch networks

