# Pouch with runV

Container technology develops rapidly recently. It provides much convenience for application packing and resource utilization improvement. At the same time it brings benefits, LXC-based container technology also loses the appropriate security. Specifically, containers share operating system kernel on one machine. Once one container attemps to attack kernel, all workload on this host would be influenced.

For some scenarios which are sensitive and strict to security, pure container technology has apparent defects. What's more, in cloud era multi-tenancy is the rigid demand from cloud customers. Therefore, strong isolation must be guaranteed seriously. In another word, container techonology needs more security hardening.

[runV](https://github.com/hyperhq/runv) is a hypervisor-based runtime for [OCI](https://github.com/opencontainers/runtime-spec). It can easily help improve host and container's security by virtualizing a guest kernel which isolates containers and host with a clearer boundary. Although hypervisor-base container brings more commitment on security, it loses some performance caused by virtualization technology. And it takes a litte more time to start a container, since more work needs done, like initrd.img loading, kernel loading, system process bootup and so on.

## Architecture

Supporting hypervisor-based OCI runtime is one of Pouch's goals. Pouch allows users to decide which kind of container to create. So with a unified entry of Pouch's API, users can create both hypervisor-based containers and LXC-based containers. With two kinds of carriers above, user's application can flexibly choose runtime on demand.

Here is the architecture of Pouch's supporting both runV and runC:

![pouch_with_runv_architecture](static_files/pouch_with_runv_architecture.png)

## Quick Start

Before installing, We should remind one important thing: **Pouch with runv can only work on PHYSICAL MACHINE**. Nested VMs currently are not supported yet. In addition, we should make sure that `containerd` and `pouchd` are already installed on the physical machine which is described in [INSTALLATION.md](../../INSTALLATION.md).

Make sure things above have been done. And then there are another three prerequisites to install before experiencing hypervisor-based containers:

* [QEMU](https://www.qemu.org): a generic machine emulator and virtualizer.
* [runv](https://github.com/hyperhq/runv): OCI-compatible runtime binary.
* [hyperstart](https://github.com/hyperhq/hyperstart): a tiny init service for hypervisor-based container.

### Install QEMU

[QEMU](https://www.qemu.org) is required to run VMs. We can execute following commands to easily install QEMU related tools.

On physical machine with Ubuntu OS installed:

```
sudo apt-get install -y qemu qemu-kvm
```

On physical machine with RedHat series OS installed:

```
sudo yum install -y qemu qemu-kvm
```

### Install runV

[runv](https://github.com/hyperhq/runv) does not provide binary package. We need to build it from source code.

First, clone version v1.0.0 of runv project from GitHub:

```
mkdir -p $Home/go/src/github.com/hyper
cd $Home/go/src/github.com/hyper
git clone --branch v1.0.0 https://github.com/hyperhq/runv.git
export GOPATH=$HOME/go
```

Second, build runv from source code:

```
sudo apt-get install autotools-dev
sudo apt-get install automake
cd runv
./autogen.sh
./configure
sudo make
sudo make install
```

Then binary runv will be located in your PATH.

### Install hyperstart

[hyperstart](https://github.com/hyperhq/hyperstart) provides init task for hypervisor-based containers. We need to build guest kernel and initrd.img from source code version v1.0.0 as well:

```
cd $Home/go/src/github.com/hyper
git clone --branch v1.0.0 https://github.com/hyperhq/hyperstart.git
cd hyperstart
./autogen.sh
./configure
sudo make
```

After building kernerl and initrd.img successfully, we should copy guest `kernel` and `initrd.img` to the default directory which runv will look for.

```
mkdir /var/lib/hyper/
cp build/{kernel,hyper-initrd.img} /var/lib/hyper/
```

### Configure containerd

With runv related tools installed, we still need to finish the last step to create hypervisor-based containers. That is to configure containerd to use runV as the default runtime option.

First, generate dafault configuration for containerd. Create file `config.toml` if it does not exist and generate config:

```
containerd config default > /etc/containerd/config.toml
```

Second, append runv as a runtime plugin config to file `config.toml`:

```
[plugins.linux]
 shim = "containerd-shim"
 no_shim = false
 runtime = "runv"
 shim_debug = true
```

### Create hypervisor-based container

After starting `pouchd`, we can test with command line tool `pouch` binary:

```
$ pouch create docker.io/library/busybox:latest
container ID: a07ae55306d276b9627ecf612bae47509e00e37afd3765b4c091f3e865271cfa, name: a07ae5 

$ pouch start -i a07ae5
/ # 
```
