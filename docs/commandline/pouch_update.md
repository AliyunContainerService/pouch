## pouch update

Update the configurations of a container

### Synopsis

Update the configurations of a container

```
pouch update [OPTIONS] CONTAINER
```

### Options

```
      --blkio-weight uint16    Block IO (relative weight), between 10 and 1000, or 0 to disable
      --cpu-share int          CPU shares (relative weight)
      --cpuset-cpus string     CPUs in cpuset
      --cpuset-mems string     MEMs in cpuset
  -e, --env strings            Set environment variables for container
  -h, --help                   help for update
      --image string           Image of container
  -l, --label strings          Set label for container
  -m, --memory string          Container memory limit
      --memory-swap string     Container swap limit
      --memory-wappiness int   Container memory swappiness [0, 100] (default -1)
      --restart string         Restart policy to apply when container exits
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

