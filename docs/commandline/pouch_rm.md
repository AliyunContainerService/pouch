## pouch rm

Remove one or more containers

### Synopsis


Remove a container object in Pouchd.
If a container be stopped or created, you can remove it. 
If the container be running, you can also remove it with flag force.
When the container be removed, the all resource of the container will
be released.


```
pouch rm [OPTIONS] CONTAINER [CONTAINER...]
```

### Examples

```
$ pouch rm 5d3152
5d3152

$ pouch rm -f 493028
493028
```

### Options

```
  -f, --force   if the container is running, force to remove it
  -h, --help    help for rm
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

