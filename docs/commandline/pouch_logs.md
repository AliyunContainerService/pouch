## pouch logs

Print a container's logs

### Synopsis

Print a container's logs

```
pouch logs [OPTIONS] CONTAINER
```

### Options

```
      --details        Show extra provided to logs
  -f, --follow         Follow log output
  -h, --help           help for logs
      --since string   Show logs since timestamp
      --tail string    Number of lines to show from the end of the logs default "all" (default "all")
  -t, --timestamps     Show timestamps
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

