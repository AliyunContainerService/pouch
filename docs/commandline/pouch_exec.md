## pouch exec

Run a command in a running container

### Synopsis

Run a command in a running container

```
pouch exec [OPTIONS] CONTAINER COMMAND [ARG...]
```

### Examples

```
$ pouch exec -it 25bf50 ps
PID   USER     TIME  COMMAND
    1 root      0:00 /bin/sh
   38 root      0:00 ps

```

### Options

```
  -d, --detach        Run the process in the background
  -h, --help          help for exec
  -i, --interactive   Open container's STDIN
  -t, --tty           Allocate a tty device
  -u, --user string   Username or UID (format: <name|uid>[:<group|gid>])
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

