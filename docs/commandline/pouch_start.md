## pouch start

Start a created or stopped container

### Synopsis

Start a created or stopped container

```
pouch start [container] [flags]
```

### Examples

```
# pouch start ${containerID} -a -i		
/ # ls /		
bin   dev   etc   home  proc  root  run   sys   tmp   usr   var		
/ # exit
```

### Options

```
  -a, --attach        Attach container's STDOUT and STDERR
  -h, --help          help for start
  -i, --interactive   Attach container's STDIN
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

