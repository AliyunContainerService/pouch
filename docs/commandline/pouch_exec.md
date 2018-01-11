## pouch exec

Exec a process in a running container

### Synopsis

Exec a process in a running container

```
pouch exec [OPTIONS] CONTAINER COMMAND [ARG...]
```

### Options

```
  -d, --detach        Run the process in the background
  -h, --help          help for exec
  -i, --interactive   Open container's STDIN
  -t, --tty           Allocate a tty device
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

