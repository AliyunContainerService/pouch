## pouch stats

Display a live stream of container(s) resource usage statistics

### Synopsis

stats command is to display a live stream of container(s) resource usage statistics

```
pouch stats [OPTIONS] CONTAINER [CONTAINER...]
```

### Examples

```
$ pouch stats b25ae a0067
CONTAINER ID        NAME                       CPU %               MEM USAGE / LIMIT     MEM %               NET I/O             BLOCK I/O           PIDS
b25ae88e5b70        naughty_goldwasser         0.11%               2.559MiB / 15.23GiB   0.02%               7.32kB / 0B         0B / 0B             4
a00670c2bdff        xenodochial_varahamihira   0.11%               2.887MiB / 15.23GiB   0.02%               13.3kB / 0B         14.7MB / 0B         4

```

### Options

```
  -h, --help        help for stats
      --no-stream   Disable streaming stats and only pull the first result
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

