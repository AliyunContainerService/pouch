## pouch build

Build an image from a Dockerfile

### Synopsis

Build an image from a Dockerfile

```
pouch build [OPTION] PATH
```

### Options

```
      --addr string             buildkitd address (default "unix:///run/buildkit/buildkitd.sock")
      --build-arg stringArray   Set build-time variables
  -h, --help                    help for build
  -t, --tag stringArray         Name and optionally a tag in the 'name:tag' format
      --target string           Set the target build stage to build
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

