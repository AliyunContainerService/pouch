# Pouch容器与kata

## 简介

Kata容器结合来自英特尔®透明容器和超runv的技术，为容器速度提供安全的虚拟机，其核心技术与runv相同，关于VM容器的详细信息，可见 [runV doc](https://github.com/alibaba/pouch/blob/master/docs/features/pouch_with_runV.md).

## 准备安装

kata 官方目前还未提供安装方式，可用的安装方法请见 [clear container project](https://github.com/clearcontainers)，更多细节请见 [kata-containers](https://github.com/kata-containers/community#users)。

### 安装

1. 安装qemu

运行虚拟机需要 [QEMU](https://www.qemu.org)。可执行以下命令安装QEMU相关工具。

在Ubuntu系统的物理机器上安装命令为:

```
sudo apt-get install -y qemu qemu-kvm
```

在Red Hat系列系统的物理机器上安装命令为:

```
sudo yum install -y qemu qemu-kvm
```

2. 安装客户内核和客户镜像

[kata-containers/osbuilder](https://github.com/kata-containers/osbuilder) 提供了创建客户镜像的工具，见 [detail steps](https://github.com/kata-containers/osbuilder#usage)。 但该工具未提供构建客户内核的方法，详细步骤可参考 [clearcontainers/osbuilder](https://github.com/clearcontainers/osbuilder#build-guest-kernel)。

3. 安装kata-runtime

该过程需要安装三个二进制库 [kata-runtime](https://github.com/kata-containers/runtime), [kata-proxy](https://github.com/kata-containers/proxy) 和 [kata-shim](https://github.com/kata-containers/shim), 在运行kata容器时，kata-runtime会调用kata-proxy和kata-shim。
可以很容易从源码中获取二进制库，以kata runtime为例，从github克隆代码，然后生成。

```shell
git clone https://github.com/kata-containers/runtime.git
cd runtime
make
```

### 配置kata runtime

Kata runtime从配置文件中读取配置，默认路径为 `/etc/kata-containers/configuration.toml`。
获取默认的配置文件：

```shell
git clone https://github.com/kata-containers/runtime.git
cd runtime
make
```

文件生成在 `cli/config/configuration.toml`，将生成的文件复制到默认路径下

```shell
cp cli/config/configuration.toml /etc/kata-containers/configuration.toml
```

可能需要修改配置文件，确保所有二进制文件在系统中的路径正确。

### 启动kata容器

完成所有步骤，就可以玩kata容器啦。

```shell
$ pouch run -d --runtime=kata-runtime 8ac48589692a top
00d1f38250fc76b5e66e7fa05a41d342d1b48202d24e2dbf06b20a113b2a008c

$ pouch ps
Name     ID       Status         Created         Image                              Runtime
00d1f3   00d1f3   Up 5 seconds   7 seconds ago   docker.io/library/busybox:latest   kata-runtime
```

进入kata容器。

```shell
$ pouch exec -it 00d1f3 sh
/ # uname -r
4.9.47-77.container
```
