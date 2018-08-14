## pouch checkpoint ls

list checkpoints of a container

### Synopsis

List a container checkpoint.

```
pouch checkpoint ls [OPTIONS] CONTAINER
```

### Examples

```
$ pouch checkpoint list container-name
cp0
```

### Options

```
      --checkpoint-dir string   directory to store checkpoints images
  -h, --help                    help for ls
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

* [pouch checkpoint](pouch_checkpoint.md)	 - Manage checkpoint commands

