## pouch logs

Print a container's logs

### Synopsis

Get container's logs

```
pouch logs [OPTIONS] CONTAINER
```

### Examples

```
$ pouch ps 
Name     ID       Status      Created      Image                                          Runtime
073f29   073f29   Up 1 day    2 days ago   registry.hub.docker.com/library/redis:latest   runc
$ pouch logs 073f29
1:C 04 Sep 05:42:01.600 # oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
1:C 04 Sep 05:42:01.601 # Redis version=4.0.11, bits=64, commit=00000000, modified=0, pid=1, just started
1:C 04 Sep 05:42:01.601 # Warning: no config file specified, using the default config. In order to specify a config file use redis-server /path/to/redis.conf
1:M 04 Sep 05:42:01.601 * Increased maximum number of open files to 10032 (it was originally set to 1024).
1:M 04 Sep 05:42:01.602 * Running mode=standalone, port=6379.
1:M 04 Sep 05:42:01.602 # WARNING: The TCP backlog setting of 511 cannot be enforced because /proc/sys/net/core/somaxconn is set to the lower value of 128.
1:M 04 Sep 05:42:01.602 # Server initialized
1:M 04 Sep 05:42:01.602 # WARNING overcommit_memory is set to 0! Background save may fail under low memory condition. To fix this issue add 'vm.overcommit_memory = 1' to /etc/sysctl.conf and then reboot or run the command 'sysctl vm.overcommit_memory=1' for this to take effect.
1:M 04 Sep 05:42:01.602 # WARNING you have Transparent Huge Pages (THP) support enabled in your kernel. This will create latency and memory usage issues with Redis. To fix this issue run the command 'echo never > /sys/kernel/mm/transparent_hugepage/enabled' as root, and add it to your /etc/rc.local in order to retain the setting after a reboot. Redis must be restarted after THP is disabled.
1:M 04 Sep 05:42:01.602 * Ready to accept connections
```

### Options

```
      --details        Show extra details provided to logs
  -f, --follow         Follow log output
  -h, --help           help for logs
      --since string   Show logs since timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)
      --tail string    Number of lines to show from the end of the logs default "all" (default "all")
  -t, --timestamps     Show timestamps
      --until string   Show logs before timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)
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

