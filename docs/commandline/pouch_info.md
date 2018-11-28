## pouch info

Display system-wide information

### Synopsis

Display the information of pouch, including Containers, Images, Storage Driver, Execution Driver, Logging Driver, Kernel Version, Operating System, CPUs, Total Memory, Name, ID.

```
pouch info [OPTIONS]
```

### Examples

```
$ pouch info
Containers: 1
 Running: 1
 Paused: 0
 Stopped: 0
Images:  0
ID:
Name:
Server Version: 0.3-dev
Storage Driver:
Driver Status: []
Logging Driver:
Cgroup Driver:
runc: <nil>
containerd: <nil>
Security Options: []
Kernel Version: 3.10.0-693.17.1.el7.x86_64
Operating System:
OSType: linux
Architecture:
HTTP Proxy: http://127.0.0.1:5678
HTTPS Proxy:
Registry: https://index.docker.io/v1/
Experimental: false
Debug: true
Labels: []
CPUs: 0
Total Memory: 0
Pouch Root Dir: /var/lib/pouch
LiveRestoreEnabled: false
Daemon Listen Addresses: [unix:///var/run/pouchd.sock]

```

### Options

```
  -h, --help   help for info
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

