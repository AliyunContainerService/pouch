## pouch network disconnect

Disconnect a container from a network

### Synopsis

Disconnect a container from a network

```
pouch network disconnect [OPTIONS] NETWORK CONTAINER
```

### Examples

```
$ pouch network disconnect bridge test
	
```

### Options

```
  -f, --force   Force the container to disconnect from a network
  -h, --help    help for disconnect
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

