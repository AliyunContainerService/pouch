## pouch network list

List pouch networks

### Synopsis

List networks in pouchd. It lists the network's Id, name, driver and scope.

```
pouch network list
```

### Examples

```
$ pouch network list
NETWORK ID   NAME     DRIVER   SCOPE
058fce03b8   none     null     local
b05a9b8844   bridge   bridge   local
d8684bf988   host     host     local

```

### Options

```
  -h, --help   help for list
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

