## pouch push

Push an image to registry

### Synopsis

Push a local image to remote registry.

```
pouch push IMAGE[:TAG]
```

### Examples

```
$ pouch push docker.io/testing/busybox:1.25
docker.io/testing/busybox:1.25:                                                   resolved |++++++++++++++++++++++++++++++++++++++|
manifest-sha256:29f5d56d12684887bdfa50dcd29fc31eea4aaf4ad3bec43daf19026a7ce69912: done
layer-sha256:56bec22e355981d8ba0878c6c2f23b21f422f30ab0aba188b54f1ffeff59c190:    done
config-sha256:e02e811dd08fd49e7f6032625495118e63f597eb150403d02e3238af1df240ba:   done
elapsed: 0.0 s                                                                    total:   0.0 B (0.0 B/s)

```

### Options

```
  -h, --help   help for push
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

