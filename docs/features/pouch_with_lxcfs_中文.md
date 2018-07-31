# Pouch容器和LXCFS

容器技术提供了不同于传统虚拟机技术（例如VMWare、KVM）的环境隔离方式。通常的Linux容器对容器打包和启动进行了加速，但这也降低了隔离的强度。其中Linux容器最为知名的问题就是资源视图问题。

容器方案让用户可以限制各容器资源的使用，包括内存资源、CPU资源、blkio等。容器内的进程无法访问超过预设阈值的资源，然而需要注意的是，如果容器内的一个进程使用监测资源上限的命令，如：`free`, `cat /proc/meminfo`, `cat /proc/cpuinfo`, `cat /proc/uptime`，那么这个进程看到的数据是物理机的数据，而非容器数据。

例如，在一台内存为2G的机器上创建一个容器，并将它内存上限设为200M，我们可以看到`free`命令获取的结果是机器的数据，而非容器数据：

```
$ pouch run -m 200m registry.hub.docker.com/library/ubuntu:16.04 free -h
              total        used        free      shared  buff/cache   available
Mem:           2.0G        103M        1.2G        3.3M        684M        1.7G
Swap:          2.0G          0B        2.0G
```

## 资源视图隔离的场景

如果缺乏资源视图隔离，容器内应用可能无法正常在容器提供的环境中运行。从应用的视角看，其运行时环境会和平时的物理机或者虚拟机不同。下面列出了一些这种情况下导致的应用安全隐患：

> 对于很多基于JVM的Java应用而言，应用启动脚本会很大程度上根据系统资源上限来分配JVM的堆和栈的大小。而运行在2G内存的机器上的一个容器可能只有200M内存上限，那么在这个容器里的应用可能会误以为自己有2G的内存可以支配，Java启动脚本也会因此让Java运行时以2G的内存上限为依据进行JVM堆和栈的分配。在这种情况下，应用必然会启动失败。并且在Java应用里，一些Java库也会根据资源视图分配堆和栈的大小，这同样存在安全隐患。

实际上，如果资源视图不能被合理隔离，不仅是内存上会引发安全问题，CPU上也会有安全问题。

> 大多数的中间件软件都会根据其视图的cpuinfo设定默认线程数。因此容器使用者有责任配置好容器的cpuset，cpuset的设定会在cgroup文件中生效。但是，容器内的进程总是会从`/proc/cpuinfo`中获取到CPU核的总数，而这必然导致应用的不稳定。

资源视图隔离也会影响容器内系统级的应用。

容器可以用来对系统级应用进行打包，而系统级应用往往要通过虚拟文件系统（Virtual File System）或者`/proc`获取系统信息。如果其获取的信息不是来自容器而是机器，系统级应用将出现非预期的行为。实际操作中，除了`cpuinfo`和`meminfo`，对其他资源也需要进行视图隔离。

## 什么是LXCFS

[LXCFS](https://github.com/lxc/lxcfs)是一个小型的[FUSE filesystem](https://en.wikipedia.org/wiki/Filesystem_in_Userspace)，而实现它的初衷则是让Linux容器看上去更像是一台虚拟机。最初的LXCFS只是一个LXC附属的小项目，但实际上LXCFS能被任何运行时（runtime）使用。LXCFS与Linux内核2.6+兼容，LXCFS会处理好在`procfs`中的重要信息，储存这些信息的文件包括：
 *  /proc/cpuinfo
 *  /proc/diskstats
 *  /proc/meminfo
 *  /proc/stat
 *  /proc/swaps
 *  /proc/uptime

早期版本的Pouch容器已经能很稳定地支持LXCFS，如果用户使用了LXCFS，一个对应的守护进程（daemon process）——`lxcfs`便会在主机上运行。通常来讲，创建一个有限资源的容器时，系统会在cgroup文件系统里创建一系列的映射至该容器的虚拟文件。LXCFS会动态地读取这些文件中的值（如`memory.limit_in_bytes`）并产生一系列的新的虚拟文件在主机上（如`/var/lib/lxc/lxcfs/proc/meminfo`），随后再把这些文件和容器绑定。最后，容器中的进程便可以通过读文件（如`/proc/meminfo`文件）的形式获取到正确的资源视图。

LXCFS和容器的架构图：
![pouch_with_lxcfs](../static_files/pouch_with_lxcfs.png)

## 如何开始使用

开启LXCFS后，对于用户而言，资源视图隔离实际上是透明的。其实，我们在安装Pouch容器时，安装程序会检查LXCFS是否在`$PATH`中，如果不在则会自动安装LXCFS。

在体验LXCFS的资源视图隔离前，用户需要保证在pouchd中已经将LXCFS模式开启。如果用户未曾开启LXCFS模式，则需要重启pouchd并在启动时使用`pouchd --enable-lxcfs`命令。只有启动了LXCFS模式的pouchd才能保证用户能够正常使用LXCFS的功能。

LXCFS模式开启时，pouchd才能够创建隔离了资源视图的容器，但这并不影响pouchd创建普通容器（无资源视图隔离）。

最后，对于在已经启动了LXCFS的pouchd而言，若要使得其中的容器使用LXCFS的功能，唯一的办法就是在命令`pouch run`中，添加一个`--enableLxcfs`标识。下面我们将会尝试在2G内存的主机创建一个200M内存限制的容器。

### 准备工作

确保LXCFS服务已经启动（下面的命令仅供Centos系统参考，其他系统可能需要其他命令）:

```
$ systemctl start lxcfs
$ ps -aux|grep lxcfs
root     1465765  0.0  0.0  95368  1844 ?        Ssl  11:55   0:00 /usr/bin/lxcfs /var/lib/lxcfs/
root     1465971  0.0  0.0 112736  2408 pts/0    S+   11:55   0:00 grep --color=auto lxcfs
```

启动pouchd LXCFS（使用`--enable-lxcfs`标识）：

```
$ cat /usr/lib/systemd/system/pouch.service
[Unit]
Description=pouch

[Service]
ExecStart=/usr/local/bin/pouchd --enable-lxcfs
...

$ systemctl daemon-reload && systemctl restart pouch
```

```shell
$ pouch run -m 200m --enableLxcfs registry.hub.docker.com/library/ubuntu:16.04 free -h
              total        used        free      shared  buff/cache   available
Mem:           200M        876K        199M        3.3M         12K        199M
Swap:          2.0G          0B        2.0G
```

我们可以看到，容器中总内存的大小为200M，符合我们的设定的内存上限。

执行了上面的这些命令后，我们会发现容器中的进程的资源视图看到的确确实实是我们设定的上限，这使得容器里的应用程序变得更加稳定和安全，这也是Pouch容器必不可少的功能之一。