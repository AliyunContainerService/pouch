## pouch history

Display history information on image

### Synopsis

Return the history information about image

```
pouch history [OPTIONS] IMAGE
```

### Examples

```
pouch history busybox:latest
IMAGE          CREATED      CREATED BY                                      SIZE        COMMENT
e1ddd7948a1c   1 week ago   /bin/sh -c #(nop)  CMD ["sh"]                   0.00 B
<missing>      1 week ago   /bin/sh -c #(nop) ADD file:96fda64a6b725d4...   716.06 KB  
```

### Options

```
  -h, --help       help for history
      --human      Print information in human readable format (default true)
      --no-trunc   Do not truncate output
  -q, --quiet      Only show image numeric ID
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

