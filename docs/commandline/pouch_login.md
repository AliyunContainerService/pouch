## pouch login

Login to a registry

### Synopsis


login to a v1/v2 registry with the provided credentials.

```
pouch login [OPTIONS] [SERVER]
```

### Examples

```
$ pouch login -u $username -p $password
Login Succeeded
```

### Options

```
  -h, --help              help for login
  -p, --password string   password for registry
  -u, --username string   username for registry
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

