## pouch network inspect

Inspect one or more pouch networks

### Synopsis

Inspect a network in pouchd. It must specify network's name.

```
pouch network inspect [OPTIONS] Network [Network...]
```

### Examples

```
$ pouch network inspect net1
Name:         net1
Scope:        
Driver:       bridge
EnableIPV6:   false
ID:           c33c2646dc8ce9162faa65d17e80582475bbe53dc70ba0dc4def4b71e44551d6
Internal:     false
```

### Options

```
  -f, --format string   Format the output using the given go template
  -h, --help            help for inspect
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

