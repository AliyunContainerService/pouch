# Pouch run with runV guide

You can make VMs run like containers if you combine Pouch with runV.

## What is runV

runV is a hypervisor-based implement for OCI runtime. With runV, you can easily
boot a VM with speed of container.

## What will pouch provide with runV

Pouch work with runV will provide a container which has the security of VM and
the fast boot speed as container. Traditionally, VMs get heavy dependency
and are slow to boot(seconds), through containers get the fast boot speed but not
have good solutions for resource isolation and secury. Using pouch, you can enjoy
the advantages of both vm and container.


## Prerequisites

### Install qemu

[qemu](https://www.qemu.org) is required to run VMs.

#### On ubuntu

```
sudo apt-get install -y qemu qemu-kvm
```

### Install runV

[runv](https://github.com/hyperhq/runv) does not provide binary package, build
runv from source code.

1. download runv from github
```
export GOPATH=$HOME/go
mkdir -p $GOPATH
go get -u github.com/hyperhq/runv
```

2. build runv
```
cd $GOPATH/src/github.com/hyperhq/runv
./autogen.sh
./configure
sudo make
sudo make install
```

3. install [hyperstart](https://github.com/hyperhq/hyperstart) to provide guest
kernel and initrd
```
git clone https://github.com/hyperhq/hyperstart.git
cd hyperstart
./autogen.sh
./configure
sudo make
```
