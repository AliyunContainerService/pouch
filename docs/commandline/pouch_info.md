## pouch info

Display system-wide information

### Synopsis

Display system-wide information

```
pouch info [OPTIONS]
```

### Examples

```
$ pouch info
ID:
Name:
OperatingSystem:
PouchRootDir:         /var/lib/pouch
ServerVersion:        0.3-dev
ContainersRunning:    0
Debug:                false
DriverStatus:         []
Labels:               []
Containers:           0
DefaultRuntime:       runc
Driver:
ExperimentalBuild:    false
KernelVersion:        3.10.0-693.11.6.el7.x86_64
OSType:               linux
CgroupDriver:
ContainerdCommit:     <nil>
ContainersPaused:     0
LoggingDriver:
SecurityOptions:      []
NCPU:                 0
RegistryConfig:       <nil>
RuncCommit:           <nil>
ContainersStopped:    0
HTTPSProxy:
IndexServerAddress:   https://index.docker.io/v1/
LiveRestoreEnabled:   false
Runtimes:             map[]
Architecture:
HTTPProxy:
Images:               0
MemTotal:             0

```

### Options

```
  -h, --help   help for info
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

