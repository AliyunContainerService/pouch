# Diskquota

## What is diskquota

Diskquota is one kind of technology which is used to restrict filesystem's disk
usage. PouchContainer uses diskquota to limit filesystem disk space. We all know that
the way based on block devices could directly help limit disk space usage
easily via setting size of block device. While the way based on filesystem can
hardly do this. Diskquota is designed for limitting filesystem disk usage.
Currently PouchContainer supports diskquota which is based on graphdriver overlayfs.

Currently in underlying filesystems only ext4 and xfs support diskquota. In
addition, there are two ways to make it: **group quota** and
**project quota**.

There are two dimensions to limit disk usage:

* usage quota(block quota): setting disk usage limit for a filesystem directory(not for inode number);
* file quota(inode quota): restrict file or inode allocation.

PouchContainer only supports block quota now with no inode support temporarily.

## Diskquota in PouchContainer

Diskquota in PouchContainer relies on kernel version PouchContainer runs on. Here is a table
describing when each filesystem supports diskquota.

| | alikernel  | open kernel |
| --- | --- | --- |
| ext4 | >= 2.6.32 group quota <br> >= 4.5 project quota | >= 4.5 project quota |
| xfs (unsupport) | >= 3.10 project quota | >= 3.10 project quota|

Although each filesystem in related kernel version supports diskquota, user
still needs to install [quota-tools-4.04](https://nchc.dl.sourceforge.net/project/linuxquota/quota-tools/4.04/quota-4.04.tar.gz).
This quota tool has not packaged into PouchContainer rpm yet. We will do this in the
future.

## Get Started

There are two ways in PouchContainer for a container to get involved in underlying
filesystems:

* container rootfs (container filesystem root directory and logfile of container)
* container volume bind from host(outside of container) to inside.

Both two dimensions are covered in PouchContainer's diskquota.

### Parameter Details

Flag `--disk-quota []string` is used to restrict diskquota of container's corresponding directory. The input type is `string`.

There are four ways to identify the input format:

* rule1: `--disk-quota=10GB` : maps container rootfs and all potential volumes binded inside;
* rule2: `--disk-quota=/abc=10GB` : absolute path matching, maps only mount point `/abc` has been limited, container rootfs and any other volume haven't been limited;
* rule3: `--disk-quota=/&/abc=10G` : shared size matching, maps container rootfs and mount point `/abc` have been limited, and their total block size are 10G;
* rule4: `--disk-quota=.*=10G`: regular expression matching, maps container rootfs and each mount points have been limited to 10G independently.

Flag `--quota-id string` is used to pick an existent quota ID to specify the newly input disk quota. The input type is `string` as well.

There are three ways to identify the input format:

* `--quota-id=-1` : it means pouch daemon will assign  a unique available quota id automatically to set quota, include container rootfs;
* `--quota-id=0` or --quota-id="" or haven't set: it means pouch daemon don't assign quota id;
* `--quota-id=16777216` or more than `16777216`, it means using the specified quota id to set quota.

> 1. `rule1` can't use with `rule2/rule3/rule4` together;
> 2. `quota-id` isn't null or `0`, `--disk-quota` only can been set with one rule, or means the length of disk quota slice is `1`.
> 3. `rule3` use `&` link with directory that must be absolute path, can't be regular expression with `rule4`;
> 4. no special characters(`* & ;`) can exist in all mount points
> 5. valid `quota-id` which is larger than `16777216`
> 6. In this scenario of triggering `upgrade` interface, pouchd will remove the old container and use the new image to take place of old container's image, and create a new container which should inherit the original diskquota. Then user can pass an original `quota-id` of original container to take effect on newly created container.

The effect taken by `disk-quota` and `quota-id` is like the following sheet:

| disk-quota | quota-id(<0) | quota-id(=0) | quota-id(>0)|
| :--------: | :--------:| :--: |:--: |
| 10GB | auto gen quota-id and return，rootfs+n\*volume(total 10GB) | no setting quotaID，rootfs+n\*volume(total 10GB) | setting as input quota-id, rootfs+n\*volume(total 10GB) |
| /abc=10GB | auto gen quota-id and return，only `/abc` set to 10GB) | no setting quotaID，only `/abc` set to 10GB) | setting as input quota-id, only `/abc` set to 10GB) |
| .*=10GB | auto gen quota-id and return，rootfs+n\*volume(total 10GB) | no setting quotaID，rootfs 10GB，each volume 10GB | setting as input quota-id, rootfs+n\*volume(total 10GB) |
| /&/abc=10GB | auto gen quota-id and return, rootfs+/abc=10GB, others haven't been limited | no setting quotaID，rootfs+/abc=10GB | setting as input quota-id, rootfs+/abc=10GB |

Pouchd created local volume with disk quota if user requests to create a volume with size option. If this volume is already set a disk quota rule, then no matter what directory inside container this volume is binded to, and no matter what disk quota user adds to the inside directory again, this volume will be under the original disk quota which is set at the very beginning.

### Container Rootfs diskquota

Users can set flag `--disk-quota` for a created container's rootfs to limit
disk space usage, for example `--disk-quota 10g`. After setting this
successfully, we can see rootfs size is 10GB via command `df -h`. And it shows
that diskquota has taken effects.

```bash
$ pouch run -ti --disk-quota /=10g registry.hub.docker.com/library/busybox:latest df -h
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
