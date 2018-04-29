## pouch image inspect

Display detailed information on one or more images

### Synopsis

Return detailed information on Pouch image

```
pouch image inspect [OPTIONS] IMAGE [IMAGE...]
```

### Examples

```
$ pouch image inspect docker.io/library/busybox
{
  "CreatedAt": "2017-12-21 04:30:57",
  "Digest": "sha256:bbc3a03235220b170ba48a157dd097dd1379299370e1ed99ce976df0355d24f0",
  "ID": "bbc3a0323522",
  "Name": "docker.io/library/busybox:latest",
  "Size": 720019,
  "Tag": "latest"
}
```

### Options

```
  -f, --format string   Format the output using the given go template
  -h, --help            help for inspect
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

* [pouch image](pouch_image.md)	 - Manage image

