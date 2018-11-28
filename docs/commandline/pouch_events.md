## pouch events

Get real time events from the daemon

### Synopsis

events cli tool is used to subscribe pouchd events. We support filter parameter to filter some events that we care about or not.

```
pouch events [OPTIONS]
```

### Examples

```
$ pouch events -s "2018-08-10T10:52:05"
	2018-08-10T10:53:15.071664386-04:00 volume create 9fff54f207615ccc5a29477f5ae2234c6b804ed8aad2f0dfc0dccb0cc69d4d12 (driver=local)
2018-08-10T10:53:15.091131306-04:00 container create f2b58eb6bc616d7a22bdb89de50b3f04e2c23134accdec1a9b9a7490d609d34c (image=registry.hub.docker.com/library/centos:latest, name=test)
2018-08-10T10:53:15.537704818-04:00 container start f2b58eb6bc616d7a22bdb89de50b3f04e2c23134accdec1a9b9a7490d609d34c (image=registry.hub.docker.com/library/centos:latest, name=test)
```

### Options

```
  -f, --filter strings   Filter output based on conditions provided
  -h, --help             help for events
  -s, --since string     Show all events created since timestamp
  -u, --until string     Stream events until this timestamp
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

