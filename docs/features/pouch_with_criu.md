# PouchContainer with CRIU

## Introduction

[CRIU](https://criu.org/Main_Page) is a software tool for the Linux operating system. Using this tool, you can freeze a running application (or part of it) and checkpoint it as a collection of files on disk. With the help of CRIU, pouch can dump and keep a running container's status as a checkpoint directory, and create a new container to restore previous running status from the checkpoint directory.

### Installation

Before install CRIU, check `CONFIG_CHECKPOINT_RESTORE` is enable by your kernel. More detail check, you can visit [criu linux kernel](https://criu.org/Linux_kernel), generally, check this config is enough.

```bash
$ cat /boot/config-`uname -r` | grep CONFIG_CHECKPOINT_RESTORE
CONFIG_CHECKPOINT_RESTORE=y
```

#### Install from package

Find package belongs to your linux distributions on [criu packages](https://criu.org/Packages). If your distributions is ubuntu 16.04(xenial), you can use `apt-get` to install.

```bash
sudo apt-get update
sudo apt-get install -y criu
```

#### Installation from source

1. install CRIU on ubuntu

```bash
sudo apt-get install -y build-essential libnet1-dev libprotobuf-dev libprotobuf-c0-dev \
protobuf-c-compiler protobuf-compiler python-protobuf libnl-3-dev libcap-dev asciidoc

git clone https://github.com/checkpoint-restore/criu.git

cd criu
sudo make
sudo make install
```

2. install CRIU in centos

```bash
sudo sudo yum install -y protobuf protobuf-c protobuf-c-devel protobuf-compiler protobuf-devel protobuf-python \
pkg-config python-ipaddr libbsd iproute2 libcap-devel libnet-devel libnl3-devel asciidoc xmlto

git clone https://github.com/checkpoint-restore/criu.git

cd criu
sudo make
sudo make install
```

### Usage

1. run a container with a simple task running.

```bash
pouch run -d --name=criu busybox /bin/sh -c 'while true; do sleep 1; done'
```

2. create a checkpoint from a running container.

```bash
pouch checkpoint create --checkpoint-dir=/tmp criu cp0
```

3. create a new container, start it from a checkpoint to get previous container running status. If you are using busybox image, you should start it first and stop, since then /proc can be created in container, or criu will fail.

```bash
$ pouch  create busybox
03859210443e99c6f6026ddf0d0a7a82da9480449860e343abfa566aa38fd055

# this is for create /proc in container, if is not busybox image, start and stop is no need.
$ pouch start 03859210443e99c6f6026ddf0d0a7a82da9480449860e343abfa566aa38fd055
03859210443e99c6f6026ddf0d0a7a82da9480449860e343abfa566aa38fd055

$ pouch stop 03859210443e99c6f6026ddf0d0a7a82da9480449860e343abfa566aa38fd055
03859210443e99c6f6026ddf0d0a7a82da9480449860e343abfa566aa38fd055

$ pouch start --checkpoint-dir=/tmp --checkpoint=cp0 03859210443e99c6f6026ddf0d0a7a82da9480449860e343abfa566aa38fd055
03859210443e99c6f6026ddf0d0a7a82da9480449860e343abfa566aa38fd055
```

4. check shell process in exist in new container.

```bash
$ pouch exec 03859210443e99c6f6026ddf0d0a7a82da9480449860e343abfa566aa38fd055 ps -ef
PID   USER     TIME  COMMAND
    1 root      0:00 /bin/sh -c while true; do sleep 1; done
 1791 root      0:00 sleep 1
 1792 root      0:00 ps -ef
```
