## pouch network

Manage pouch networks

### Synopsis

Manager the networks in pouchd. It contains the functions of create/remove/list/inspect network, 'driver' is used to list drivers that pouch support. Now bridge network is supported in pouchd defaulted, it will be initialized when pouchd starting.

### Options

```
  -h, --help   help for network
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

* [pouch](pouch.md)	 - An efficient container engine
* [pouch network create](pouch_network_create.md)	 - Create a pouch network

