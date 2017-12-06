## pouch rename

Rename a container with newName

### Synopsis

Rename a container object in Pouchd. You can change the name of one container identified by it's name or ID. The container you renamed is ready to be used by it's new name.

```
pouch rename [container] [newName] [flags]
```

### Examples

```
$ pouch ps
Name     ID       Status    Image
foo      71b9c1   Running   docker.io/library/busybox:latest
$ pouch rename foo newName
$ pouch ps
Name     ID       Status    Image
newName  71b9c1   Running   docker.io/library/busybox:latest

```

### Options

```
  -h, --help   help for rename
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

