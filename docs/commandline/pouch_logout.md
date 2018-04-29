## pouch logout

Logout from a registry

### Synopsis


logout from a v1/v2 registry.

```
pouch logout [SERVER]
```

### Examples

```
$ pouch logout $registry
Remove login credential for registry: $registry
```

### Options

```
  -h, --help   help for logout
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

