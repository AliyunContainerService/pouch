# PouchContainer with log driver

The PouchContainer has supported the json, syslog log driver to help you retrieve log information from running containers. If you do not specify a log driver, the default is json-file.

You can view the type of log driver by following commands

```
$ pouch info | grep 'Logging Driver'
Logging Driver:  json-file
```

## Configure default log drivers for pouch daemon

```
$ pouchd --log-driver syslog
INFO[0000] starting containerd                           module=containerd revision=773c489c9c1b21a6d78b5c538cd395416ec50f88 version=v1.0.3
INFO[0000] loading plugin "io.containerd.content.v1.content"...  module=containerd type=io.containerd.content.v1
INFO[0000] loading plugin "io.containerd.snapshotter.v1.btrfs"...  module=containerd type=io.containerd.snapshotter.v1
...
```

You can view the default log driver for pouch daemon by following commands

```
$ pouch info | grep 'Logging Driver'
Logging Driver: syslog
```

## Configuring log drivers for containers

```
$ pouch run --log-driver syslog registry.hub.docker.com/library/centos:7 echo "hello world"
hello world
```

You can view the default log driver for running containers by following commands

```
$ pouch inspect  -f {{.HostConfig.LogConfig}} 09092c
{syslog map[]}
```