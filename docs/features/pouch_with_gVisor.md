# PouchContainer with gVisor

## Introduction

gVisor is a user-space kernel, written in Go, that implements a substantial portion of the Linux system surface. You can see [gVisor readme](https://github.com/google/gvisor).

## Prerequisites Installation

Referring to [gvisor#installation](https://github.com/google/gvisor#installation), gVisor can only run on x86_64 Linux 3.17+.

### Install gVisor

```shell
wget https://storage.googleapis.com/gvisor/releases/nightly/latest/runsc
wget https://storage.googleapis.com/gvisor/releases/nightly/latest/runsc.sha512
sha512sum -c runsc.sha512
chmod a+x runsc
sudo mv runsc /usr/local/bin
```

### Configure PouchContainer

Add runsc into config file (/etc/pouch/config.json), restart pouchd, ensure that pouchd know the specified runtime.

```json
{
    "add-runtime":{
        "runsc":{
            "path":"/usr/local/bin/runsc",
            "runtimeArgs":[
                "--debug",
                "--debug-log=/home/logs/"
            ]
        }
    }
}
```

## Start container

With all the steps finished, you can play with gVisor container.

1. start container

```shell
$ pouch run -d --runtime=runsc busybox top
822baedc4f965ed05a64524d6d91f9d4f256561f56087e926cf81fef39dde597

$ pouch ps
Name     ID       Status          Created          Image       Runtime
822bae   822bae   Up 24 seconds   25 seconds ago   busybox     runsc
```

2. enter into the gVisor container.

```shell
$ pouch  exec -it 822bae sh
/ # ls
__runsc_containers__  home                  tmp
bin                   proc                  usr
dev                   root                  var
etc                   sys
/ # uname -r
3.11.10
```