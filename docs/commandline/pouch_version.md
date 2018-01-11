## pouch version

Print versions about Pouch CLI and Pouchd

### Synopsis

Print versions about Pouch CLI and Pouchd

```
pouch version
```

### Examples

```
$ pouch version
GoVersion:       go1.9.1
KernelVersion:
Os:              linux
Version:         0.1.0-dev
APIVersion:      1.24
Arch:            amd64
BuildTime:       2017-12-18T07:48:56.348129663Z
GitCommit:

```

### Options

```
  -h, --help   help for version
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

