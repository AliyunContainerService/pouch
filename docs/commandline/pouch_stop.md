## pouch stop

Stop a running container

### Synopsis

Stop a running container in Pouchd. This is useful when you wish to stop a container.And Pouchd will stop this running container and release the resource. The container that you stopped will be closed. 

```
pouch stop [container] [flags]
```

### Examples

```
$ pouch ps
Name:foo ID:71b9c1 Status:Running Image:docker.io/library/busybox:latest 
$ pouch stop foo 
$ pouch ps 
Name:foo ID:71b9c1 Status:Stopped Image:docker.io/library/busybox:latest
```

### Options

```
  -h, --help   help for stop
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

