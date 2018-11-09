## pouch version

Print versions about Pouch CLI and Pouchd

### Synopsis

Display the version information of pouch client and daemon， including GoVersion, KernelVersion, Os, Version, APIVersion, Arch, BuildTime and GitCommit.

```
pouch version
```

### Examples

```
$ pouch version
GoVersion:       go1.10.4
KernelVersion:   3.10.0-693.11.6.el7.x86_64
Os:              linux
Version:         1.0.0
APIVersion:      1.24
Arch:            amd64
BuildTime:       2018-11-07T07:48:56.348129663Z
GitCommit:

```

### Options

```
  -h, --help   help for version
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

