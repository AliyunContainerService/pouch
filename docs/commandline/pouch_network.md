## pouch network

Manage pouch networks

### Synopsis

Manager the networks in pouchd. It contains the functions of create/remove/list/inspect network, 'driver' is used to list drivers that pouch support. Now bridge network is supported in pouchd defaulted, it will be initialized when pouchd starting.

```
pouch network [command]
```

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
* [pouch network disconnect](pouch_network_disconnect.md)	 - Disconnect a container from a network
* [pouch network inspect](pouch_network_inspect.md)	 - Inspect one or more pouch networks
* [pouch network list](pouch_network_list.md)	 - List pouch networks
* [pouch network remove](pouch_network_remove.md)	 - Remove a pouch network

