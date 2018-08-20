# PouchContainer with Diskquota

## What is diskquota

Diskquota is one kind of technology which is used to restrict filesystem's disk
usage. PouchContainer uses diskquota to limit filesystem disk space. We all know that
the way based on block devices could directly help limit disk space usage
easily via setting size of block device. While the way based on filesystem can
hardly do this. Diskquota is designed for limitting filesystem disk usage.
Currently PouchContainer supports diskquota which is based on graphdriver overlayfs.

Currently in underlying filesystems only ext4 and xfs support diskquota. In
addition, there are three ways to make it: **user quota**, **group quota** and
**project quota**.

There are two dimensions to limit disk usage:

* usage quota(block quota): setting disk usage limit for a filesystem directory(not for inode number);
* file quota(inode quota): restrict file or inode allocation.

PouchContainer only supports block quota now with no inode support temporarily.

## Diskquota in PouchContainer

Diskquota in PouchContainer relies on kernel version PouchContainer runs on. Here is a table
describing when each filesystem supports diskquota.

|| user/group quota | project quota|
|:---:| :----:| :---:|
|ext4| >= 2.6|>= 4.5|
|xfs|>= 2.6|>= 3.10|

Although each filesystem in related kernel version supports diskquota, user
still needs to install [quota-tools-4.04](https://nchc.dl.sourceforge.net/project/linuxquota/quota-tools/4.04/quota-4.04.tar.gz).
This quota tool has not packaged into PouchContainer rpm yet. We will do this in the
future.

## Get Started

There are two ways in PouchContainer for a container to get involved in underlying
filesystems. One is container rootfs, the other is container volume bind from
host(outside of container) to inside. Both two dimensions are covered in PouchContainer.

### Container Rootfs diskquota

Users can set flag `--disk-quota` for a created container's rootfs to limit
disk space usage, for example `--disk-quota 10g`. After setting this
successfully, we can see rootfs size is 10GB via command `df -h`. And it shows
that diskquota has taken effects.

```bash
$ pouch run -ti --disk-quota 10g registry.hub.docker.com/library/busybox:latest df -h
Filesystem                Size      Used Available Use% Mounted on
overlay                  10.0G     24.0K     10.0G   0% /
tmpfs                    64.0M         0     64.0M   0% /dev
shm                      64.0M         0     64.0M   0% /dev/shm
tmpfs                    64.0M         0     64.0M   0% /run
tmpfs                    64.0M         0     64.0M   0% /proc/kcore
tmpfs                    64.0M         0     64.0M   0% /proc/timer_list
tmpfs                    64.0M         0     64.0M   0% /proc/sched_debug
tmpfs                     1.9G         0      1.9G   0% /sys/firmware
tmpfs                     1.9G         0      1.9G   0% /proc/scsi
```

### Volume Diskquota

Users can also setting volume's disk quota when creating one. It is quite easy
to add a `--option` or `-o` flag to specify disk space limit to be desired
number, for example `-o size=10g`.

After creating diskquota limited volume, users can bind this volume to a
running container. In the following example, it executes command
`pouch run -ti -v volume-quota-test:/mnt registry.hub.docker.com/library/busybox:latest df -h`.
And in running container, directory `/mnt` is restricted to be size 10GB.

```bash
$ pouch volume create -n volume-quota-test -d local -o mount=/data/volume -o size=10g
Name:         volume-quota-test
Scope:
Status:       map[mount:/data/volume sifter:Default size:10g]
CreatedAt:    2018-3-24 13:35:08
Driver:       local
Labels:       map[]
Mountpoint:   /data/volume/volume-quota-test

$ pouch run -ti -v volume-quota-test:/mnt registry.hub.docker.com/library/busybox:latest df -h
Filesystem                Size      Used Available Use% Mounted on
overlay                  20.9G    212.9M     19.6G   1% /
tmpfs                    64.0M         0     64.0M   0% /dev
shm                      64.0M         0     64.0M   0% /dev/shm
tmpfs                    64.0M         0     64.0M   0% /run
/dev/sdb2                10.0G      4.0K     10.0G   0% /mnt
tmpfs                    64.0M         0     64.0M   0% /proc/kcore
tmpfs                    64.0M         0     64.0M   0% /proc/timer_list
tmpfs                    64.0M         0     64.0M   0% /proc/sched_debug
tmpfs                     1.9G         0      1.9G   0% /sys/firmware
tmpfs                     1.9G         0      1.9G   0% /proc/scsi
```
