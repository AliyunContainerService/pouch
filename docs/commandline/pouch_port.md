## pouch port

List port mappings or a specific mapping for the container

### Synopsis

Return port binding information on Pouch container

```
pouch port CONTAINER [PRIVATE_PORT[/PROTO]]
```

### Examples

```
$ pouch run -d -p 6379:6379 -p 6380:6380/udp  redis:latest
179eba2c29fb27a000bcda75cb2be271d1833ab140d1133799d0d4d865abc44e
$ pouch port 179
6379/tcp -> 0.0.0.0:6379
6380/udp -> 0.0.0.0:6380
$ pouch port 179 6379
0.0.0.0:6379
$ pouch port 179 6380/udp
0.0.0.0:6380

```

### Options

```
  -h, --help   help for port
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

