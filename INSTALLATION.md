# Quick-Start

Two quick-starts are provided, one for end-users, the other one for developers.

As an end user who wish to use Pouch, please read [End User Quick-Start](#end-user-quick-start) to install and explore Pouch.

As a developer who wish to hack on Pouch, please read [Developer Quick-Start](#developer-quick-start) to start hacking and get invloved in the project!

## End User Quick-Start

You can install Pouch automatically on your machine with very few steps. Currently we support two kinds of Linux Distribution: Ubuntu and CentOS.

### Ubuntu

To install Pouch, you need a maintained version of Ubuntu 16.04 (Xenial LTS). Archived versions aren’t supported or tested.

Pouch is conflict with Docker, so you must uninstall Docker before installing Pouch.

**Prerequisites**

Pouch supports lxcfs to provide strong isolation, so you should install lxcfs firstly. By default, lxcfs is enabled.

``` bash
sudo apt-get install lxcfs
```

Install packages to allow 'apt' to use a repository over HTTPS:

``` bash
sudo apt-get install curl apt-transport-https ca-certificates software-properties-common
```

**1. Add Pouch's official GPG key**

``` bash
curl -fsSL http://mirrors.aliyun.com/opsx/pouch/linux/debian/opsx@service.alibaba.com.gpg.key | sudo apt-key add -
```

Verify that you now have the key with the fingerprint `B615 DDD7 90C7 0912 582D  DC2D D7AE A5ED 439A E9EC`, by searching for the last 8 characters of the fingerprint.

``` bash
$ apt-key fingerprint 439AE9EC
pub   2048R/439AE9EC 2018-01-31
      Key fingerprint = B615 DDD7 90C7 0912 582D  DC2D D7AE A5ED 439A E9EC
uid                  Pouch Release <pouch-dev@list.alibaba-inc.com>
sub   2048R/B3D2A915 2018-01-31
```

**2. Set up the pouch repository**

Before you install Pouch for the first time on a new host machine, you need to set up the Pouch repository. We enabled `stabel` repository by default, you always need the `stable` repository. To add the `test` repository, add the word `test` after the word `stable` in the command line below. Afterward, you can install and update Pouch from the repository.

``` bash
sudo add-apt-repository "deb http://mirrors.aliyun.com/opsx/pouch/linux/debian/ pouch stable"
```

**3. Install pouch**

Install the latest version of Pouch.

``` bash
# update the apt package index
sudo apt-get update
sudo apt-get install pouch
```

After installing Pouch, the `pouch` group is created, but no users are added to the group.

**4. Start pouch**

``` bash
sudo service pouch start
```

Afterwards, you can pull an image and run Pouch containers.

### CentOS

To install Pouch, you need a maintained version of CentOS 7. Archived versions aren’t supported or tested.

We have put rpm package to Aliyun mirrors, you can install pouch using Pouch repository. If you install Pouch for the first on a new host machine, you need to set up the Pouch repository. Then, you can install and update Pouch from repository.

**1.Install yum-utils**

Install required packages. yum-utils provides the yum-config-manager utility.

``` bash
sudo yum install -y yum-utils
```

**2. Set up the pouch repository**

Use the following command to add Pouch repository.

``` bash
sudo yum-config-manager --add-repo http://mirrors.aliyun.com/opsx/opsx-centos7.repo
sudo yum update
```

Note: The above command set up the `stable` repository, you can enable `test` repository by the following command.

``` bash
sudo yum-config-manager --enable pouch-test
```

You can disable the `test` repository by running the `yum-config-manager` command with the `--disable` flag. To re-enable it, use the `--enable` flag. The following command disables the test repository.

``` bash
sudo yum-config-manager --disable pouch-test
```

**3. Install pouch**

Run the following command to install the latest version of Pouch. If it's the first time to install Pouch on your host, you will be prompted to accept the GPG key, and the key's fingerprint will be shown.

``` bash
sudo yum install pouch
```

After installing Pouch, the `pouch` group is created, but no users are added to the group.

**4. Start pouch**

``` bash
sudo systemctl start pouch
```

Afterwards, you can pull an image and run Pouch containers.

## Uninstall pouch

On Ubuntu

``` bash
sudo apt-get purge pouch
```

On CentOS

``` bash
sudo yum remove pouch
```

After running the `remove` command, images, containers, volumes, or customized configuration files on your host are not automatically removed. To delete all images, containers and volumes, execute the following command:

``` bash
sudo rm -rf /var/lib/pouch
```

## Developer Quick-Start

This guide provides step by step instructions to deploy Pouch on bare metal servers or virtual machines.
As a developer, you may need to build and test Pouch binaries via source code. To build pouchd which is so-called "pouch daemon" and pouch which is so-called "pouch cli", the following build system dependencies are required:

* Linux Kernel 3.10+
* Go 1.9.0+
* containerd: 1.0.0
* runc: 1.0.0-rc4
* runv: 1.0.0 (option)

### Prerequisites Installation

Since pouchd is a kind of container engine, and pouch is a cli tool, if you hope to experience container management ability via Pouch, there are several additional binaries needed:

* [containerd](https://github.com/containerd/containerd): an industry-standard container runtime;
* [runc](https://github.com/opencontainers/runc): a CLI tool for spawning and running containers according to the OCI specification;
* [runv](https://github.com/hyperhq/runv): a hypervisor-based runtime for OCI.

Here are the shell scripts to install `containerd` and `runc`:

``` shell
# install containerd
$ wget https://github.com/containerd/containerd/releases/download/v1.0.0/containerd-1.0.0.linux-amd64.tar.gz
$ tar -xzvf containerd-1.0.0.linux-amd64.tar.gz -C /usr/local
$
# install runc
$ wget https://github.com/opencontainers/runc/releases/download/v1.0.0-rc4/runc.amd64 -P /usr/local/bin
$ chmod +x /usr/local/bin/runc.amd64
$ mv /usr/local/bin/runc.amd64 /usr/local/bin/runc
```

### runV Installation

If you wish to experience hypervisor-based virtualization additionally, you will still need to install [runV](https://github.com/hyperhq/runv).

More guide on experiencing Pouch with runV including runv Installation, please refer to [pouch run with runv guide](docs/features/pouch_with_runV.md).

### Pouch Build and Installation

With all prerequisites installed, you can build and install pouch daemon and pouch cli. Clone the repository and checkout whichever branch you like (in the following example, checkout branch master):

``` shell
mkdir -p $GOPATH/src/github.com/alibaba/
cd $GOPATH/src/github.com/alibaba/; git clone https://github.com/alibaba/pouch.git
cd pouch; git checkout master
```

Makefile target named `build` will compile the pouch and pouchd binaries in current work directory. Or you can just execute `make install` to build binaries and install them in destination directory (`/usr/local/bin` by default).

``` shell
make install
```

### Start Pouch

With all needed binaries installed, you could start pouchd via:

``` shell
$ pouchd
INFO[0000] starting containerd                           module=containerd revision=a543c937eb0a05e1636714ee2be70819d745b960 version=v1.0.0-beta.2
INFO[0000] setting subreaper...                          module=containerd
INFO[0000] loading plugin "io.containerd.content.v1.content"...  module=containerd type=io.containerd.content.v1
INFO[0000] loading plugin "io.containerd.snapshotter.v1.btrfs"...  module=containerd type=io.containerd.snapshotter.v1
WARN[0000] failed to load plugin io.containerd.snapshotter.v1.btrfs  error="path /var/lib/containerd/io.containerd.snapshotter.v1.btrfs must be a btrfs filesystem to be used with the btrfs snapshotter" module=containerd
INFO[0000] loading plugin "io.containerd.snapshotter.v1.overlayfs"...  module=containerd type=io.containerd.snapshotter.v1
INFO[0000] loading plugin "io.containerd.metadata.v1.bolt"...  module=containerd type=io.containerd.metadata.v1
WARN[0000] could not use snapshotter btrfs in metadata plugin  error="path /var/lib/containerd/io.containerd.snapshotter.v1.btrfs must be a btrfs filesystem to be used with the btrfs snapshotter" module="containerd/io.containerd.metadata.v1.bolt"
INFO[0000] loading plugin "io.containerd.differ.v1.walking"...  module=containerd type=io.containerd.differ.v1
INFO[0000] loading plugin "io.containerd.grpc.v1.containers"...  module=containerd type=io.containerd.grpc.v1
```

After pouchd's running, you could interact with pouchd by pouch cli:

```bash
$ pouch images
IMAGE ID             IMAGE NAME                                               SIZE
3e8fa85ddfef         docker.io/library/busybox:latest                         2699
504cf109b492         docker.io/library/redis:alpine                           2035
```

## Feedback

We hope this guide would help you get up and run with Pouch. And feel free to send feedback via [ISSUE](https://github.com/alibaba/pouch/issues/new), if you have any questions. If you wish to contribute to Pouch on this guide, please just submit a pull request.
