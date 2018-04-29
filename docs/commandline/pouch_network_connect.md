## pouch network connect

Connect a container to a network

### Synopsis

Connect a container to a network in pouchd. It must specify network's name and container's name.

```
pouch network connect [OPTIONS] NETWORK CONTAINER
```

### Examples

```
$ pouch network connect net1 container1
container container1 is connected to network net1
```

### Options

```
      --alias strings           Add network-scoped alias for the container
  -h, --help                    help for connect
      --ip string               IP Address
      --ip6 string              IPv6 Address
      --link strings            Add link to another container
      --link-local-ip strings   Add a link-local address for the container
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

