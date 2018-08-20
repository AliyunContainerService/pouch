# PouchContainer with kata

## Introduction

Kata Containers combines technology from Intel® Clear Containers and Hyper runV to provide the speed of containers with the security of virtual machines, the core technology is same with runV, about the detail information in vm container , you can see [runV doc](https://github.com/alibaba/pouch/blob/master/docs/features/pouch_with_runV.md).

## Prerequisites Installation

kata announces that it not provide an installation option yet, so some installation methods we get from [clear container project](https://github.com/clearcontainers), for more detail, see [kata-containers](https://github.com/kata-containers/community#users).

### Installation

1. install qemu

[QEMU](https://www.qemu.org) is required to run VMs. We can execute following commands to easily install QEMU related tools.

On physical machine with Ubuntu OS installed:

```
sudo apt-get install -y qemu qemu-kvm
```

On physical machine with Red Hat series OS installed:

```
sudo yum install -y qemu qemu-kvm
```

2. Install guest kernel and guest image

[kata-containers/osbuilder](https://github.com/kata-containers/osbuilder) provide a tool to create guest image, see the [detail steps](https://github.com/kata-containers/osbuilder#usage). Since the tool is not giving method to build guest kernel, you can see the detail steps in [clearcontainers/osbuilder](https://github.com/clearcontainers/osbuilder#build-guest-kernel).

3. install kata-runtime

In this step, we need three binary to install, [kata-runtime](https://github.com/kata-containers/runtime), [kata-proxy](https://github.com/kata-containers/proxy) and [kata-shim](https://github.com/kata-containers/shim), kata-proxy and kata-shim will called by kata-runtime in running a kata container.
It is quite easy to get the binary from the source code, let's take kata runtime for example, clone code from github, then make.

```shell
git clone https://github.com/kata-containers/runtime.git
cd runtime
make
```

### Configure kata runtime

Kata runtime read config from configuration file, it default path is `/etc/kata-containers/configuration.toml`.
Get default configuration file:

```shell
git clone https://github.com/kata-containers/runtime.git
cd runtime
make
```

File will be generated in `cli/config/configuration.toml`, copy the file into default path

```shell
cp cli/config/configuration.toml /etc/kata-containers/configuration.toml
```

You might need to modify this file, make sure that all binaries have right path in system.

### Start kata container

With all the steps finish, you can play with kata container.

1. add kata-runtime into config file, restart pouchd, ensure that pouchd know the specified runtime.

```
{
    "add-runtime": {
        "kata-runtime": {
            "path": "/usr/local/bin/kata-runtime"
        }
    }
}
```

2. start kata container.

```shell
$ pouch run -d --runtime=kata-runtime 8ac48589692a top
00d1f38250fc76b5e66e7fa05a41d342d1b48202d24e2dbf06b20a113b2a008c

$ pouch ps
Name     ID       Status         Created         Image                              Runtime
00d1f3   00d1f3   Up 5 seconds   7 seconds ago   docker.io/library/busybox:latest   kata-runtime
```

3. enter into the kata container.

```shell
$ pouch exec -it 00d1f3 sh
/ # uname -r
4.9.47-77.container
```
