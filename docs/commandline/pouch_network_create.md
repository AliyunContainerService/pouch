## pouch network create

Create a pouch network

### Synopsis

Create a network in pouchd. It must specify network's name and driver. You can use 'network driver' to get drivers that pouch support.

```
pouch network create [OPTIONS] [NAME]
```

### Examples

```
$ pouch network create -n pouchnet -d bridge --gateway 192.168.1.1 --subnet 192.168.1.0/24
pouchnet: e1d541722d68dc5d133cca9e7bd8fd9338603e1763096c8e853522b60d11f7b9
```

### Options

```
  -d, --driver string        the driver of network (default "bridge")
      --enable-ipv6          enable ipv6 network
      --gateway string       the gateway of network
  -h, --help                 help for create
      --ip-range string      the range of network's ip
      --ipam-driver string   the ipam driver of network (default "default")
  -l, --label strings        create network with labels
  -n, --name string          the name of network
  -o, --option strings       create network with options
      --subnet string        the subnet of network
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

* [pouch network](pouch_network.md)	 - Manage pouch networks

