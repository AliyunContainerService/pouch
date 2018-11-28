## pouch updatedaemon

Update the configurations of pouchd

### Synopsis

Update daemon's configurations, if daemon is stoped, it will just update config file. Online update just including: image proxy, label, offline update including: manager white list, debug level, execute root directory, bridge name, bridge IP, fixed CIDR, defaut gateway, iptables, ipforwark, userland proxy. If pouchd is alive, you can only use --offline=true to update config file

```
pouch updatedaemon [OPTIONS]
```

### Examples

```
$ pouch updatedaemon --debug=true
```

### Options

```
      --bip string                  update daemon bridge IP
      --bridge-name string          update daemon bridge device
      --config-file string          specified config file for updating daemon (default "/etc/pouch/config.json")
      --default-gateway string      update daemon bridge default gateway
      --exec-root-dir string        update exec root directory for network
      --fixed-cidr string           update daemon bridge fixed CIDR
  -h, --help                        help for updatedaemon
      --image-proxy string          update daemon image proxy
      --ipforward                   udpate daemon with ipforward (default true)
      --iptables                    update daemon with iptables (default true)
      --label strings               update daemon labels
      --manager-white-list string   update daemon manager white list
      --offline                     just update daemon config file
      --userland-proxy              update daemon with userland proxy
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

