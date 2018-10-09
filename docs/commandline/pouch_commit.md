## pouch commit

Commit an image from a container

### Synopsis

commit an image from a container.

```
pouch commit [OPTIONS] CONTAINER REPOSITORY[:TAG]
```

### Examples

```
$ pouch commit 25bf50 test:image
1c7e415csa333

```

### Options

```
  -a, --author string    Image author, eg.(name <email@email.com>)
  -h, --help             help for commit
  -m, --message string   Commit message
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

