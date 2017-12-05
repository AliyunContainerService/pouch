## pouch volume create

Create a volume

### Synopsis

Create a volume

```
pouch volume create [args] [flags]
```

### Options

```
  -d, --driver string      Specify volume driver name (default 'local') (default "local")
  -h, --help               help for create
  -l, --label strings      Set labels for volume
  -n, --name string        Specify name for volume
  -o, --option strings     Set volume driver options
  -s, --selector strings   Set volume selectors
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

* [pouch volume](pouch_volume.md)	 - Manage pouch volumes

