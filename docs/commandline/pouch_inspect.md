## pouch inspect

Get the detailed information of container

### Synopsis

Return detailed information on Pouch container

```
pouch inspect CONTAINER
```

### Examples

```
$ pouch inspect 08e
{
  "Id": "08ee444faa3c6634ecdecea26de46e8a6a16efefd9afb72eb3457320b333fc60",
  "Created": "2017-12-04 14:48:59",
  "Path": "",
  "Args": null,
  "State": {
    "StartedAt": "0001-01-01T00:00:00Z",
    "Status": 0,
    "FinishedAt": "0001-01-01T00:00:00Z",
    "Pid": 25006,
    "ExitCode": 0,
    "Error": ""
  },
  "Image": "registry.docker-cn.com/library/centos:latest",
  "ResolvConfPath": "",
  "HostnamePath": "",
  "HostsPath": "",
  "LogPath": "",
  "Name": "08ee44",
  "RestartCount": 0,
  "Driver": "",
  "MountLabel": "",
  "ProcessLabel": "",
  "AppArmorProfile": "",
  "ExecIDs": null,
  "HostConfig": null,
  "HostRootPath": ""
}
```

### Options

```
  -h, --help   help for inspect
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

