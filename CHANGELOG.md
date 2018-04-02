# CHANGELOG

## 0.3.0 (2018-03-30)

**IMPORTANT**: Pouch 0.3.0 has met almost all your basic needs for kubernetes:

1. Sandbox/Container lifecycle management
2. Image management
3. Network management with CNI
4. Container streaming: exec/attach/portforward
5. Container logging
6. Security Context: RunAsUser, Apparmor,Seccomp,Sysctl
7. Validation test of cri-tools: 51/55 Pass

**IMPORTANT**:
Kubernetes 1.10 was released recently and the CRI of it has updated from v1alpha1 to v1alpha2.
Pouch will first try to fully support CRI v1alpha1 and then CRI v1alpha2.
So kubernetes 1.9.X is recommended now

### Remote API && Client

* `inspect` now support input multi items [\#989](https://github.com/alibaba/pouch/pull/989)
* Refactor make route code much more simplified [\#988](https://github.com/alibaba/pouch/pull/988)
* Fix `restart` router miss versionMatcher [\#986](https://github.com/alibaba/pouch/pull/986)
* Add kernel value in `pouch version` command [\#942](https://github.com/alibaba/pouch/pull/942)
* Make `pouch info` print more pretty [\#941](https://github.com/alibaba/pouch/pull/941)
* Add `no-trunc` flag to `pouch ps` [\#909](https://github.com/alibaba/pouch/pull/909)
* New `restart` command that allow restarting an running container [\#890](https://github.com/alibaba/pouch/pull/890)
* New `upgrade` command that allow upgrading image and resources of a container [\#852](https://github.com/alibaba/pouch/pull/852)
* New `top` command that allow showing processes informations in container [\#878](https://github.com/alibaba/pouch/pull/878)
* Add `--format` flag to `pouch image inspect` and `pouch network inspect` commands [\#871](https://github.com/alibaba/pouch/pull/871)
* New `pouch info` command to print all informations about th pouch daemon [\#859](https://github.com/alibaba/pouch/pull/859)
* New `pouch logs` command that allow printing logs of container [\#886](https://github.com/alibaba/pouch/pull/886)
* Using the default registry when execute `logout`command if not specified one [\#902](https://github.com/alibaba/pouch/pull/902)
* New `resize` command that allow changing the height and width of TTY of an running container [\#879](https://github.com/alibaba/pouch/pull/879)

### Pouch Daemon

* New `update` API that allow updating `labels` and `image-proxy` parameters of pouch daemon [\#987](https://github.com/alibaba/pouch/pull/987)
* Add `--label` flag to `pouchd` [\#982](https://github.com/alibaba/pouch/pull/982)

### Runtime

* Change container parameter's json name from `ID` to `Id` to be compatible with Moby API [\#1002](https://github.com/alibaba/pouch/pull/1002)
* Fix errors checked by markdownlint [\#974](https://github.com/alibaba/pouch/pull/974)
* Refactor the pouch ctrd layer client interface [\#971](https://github.com/alibaba/pouch/pull/971)
* Refactor the `pkg` package that let's client package independent to other inner pkgs [\#962](https://github.com/alibaba/pouch/pull/962)
* Add circleci to pouch to validate markdown files [\#961](https://github.com/alibaba/pouch/pull/961)
* Fix log initialization of `libnetwork` in pouch [\#956](https://github.com/alibaba/pouch/pull/956)
* Fix the import path of package `logrus` [\#953](https://github.com/alibaba/pouch/pull/953)
* Fix should checking `kernelVersion.Kernel` not `kernelVersion.Major` when setting disk quota driver [\#946](https://github.com/alibaba/pouch/pull/946)
* New `restart` interface that allow restarting an running container [\#944](https://github.com/alibaba/pouch/pull/944)
* Refactor modify logic in complement image fullname [\#940](https://github.com/alibaba/pouch/pull/940)
* Add `--oom-kill-disable` and `--oom-score-adj` flags to `pouch create` [\#934](https://github.com/alibaba/pouch/pull/934)
* New `resize` API that allow changing the height and width of TTY of an running container [\#931](https://github.com/alibaba/pouch/pull/931)
* Fix execute `pouch images` command panic when pulling an image failed before [\#926](https://github.com/alibaba/pouch/pull/926)
* New `upgrade` API that allow upgrading the image and resource of a container [\#923](https://github.com/alibaba/pouch/pull/923)
* New plugin framework to support executing custom codes at plugin points [\#919](https://github.com/alibaba/pouch/pull/919)
* Add default registry namespace [\#911](https://github.com/alibaba/pouch/pull/911)
* New `top` API that allow showing the processes informations in an running container [\#900](https://github.com/alibaba/pouch/pull/900)
* Fix `cgroup-parent` should always be abs [\#896](https://github.com/alibaba/pouch/pull/896)
* Refactor set lxcfs service managed by systemd [\#885](https://github.com/alibaba/pouch/pull/885)
* Add version information in restful api url [\#869](https://github.com/alibaba/pouch/pull/869)
* Add `repoTags` and `repoDigests` in `ImageInfo` struct [\#721](https://github.com/alibaba/pouch/pull/721)

### Documentation

* Add introduction document to diskquota [\#972](https://github.com/alibaba/pouch/pull/972)
* Improve test guidance doc [\#856](https://github.com/alibaba/pouch/pull/856)

### Storage

* Add `--volume` flag to `pouch create` to support bind mounts for files [\#937](https://github.com/alibaba/pouch/pull/937)
* Fix volume can be removed when using by container [\#888](https://github.com/alibaba/pouch/pull/888)
* Add disk quota for container's rootfs [\#876](https://github.com/alibaba/pouch/pull/876)

### Kubernetes

* With this PR, we can get the error informations when stream server handles `exec` or `attach` commands occured errors [\#1007](https://github.com/alibaba/pouch/pull/1007)
* Add websocket support for cri stream server [\#985](https://github.com/alibaba/pouch/pull/985)
* Fix handle image format 'namespace/name:tag' correctly [\#981](https://github.com/alibaba/pouch/pull/981)
* Fix pull image and get its status with RefDigest [\#973](https://github.com/alibaba/pouch/pull/973)
* Store sandbox config informations for cri manager [\#955](https://github.com/alibaba/pouch/pull/955)
* Seperate stdout & stderr of container io and support host network mode for sandbox [\#945](https://github.com/alibaba/pouch/pull/945)
* Implement ReadOnlyRootfs and add `no-new-privilegs` support to cri manager [\#935](https://github.com/alibaba/pouch/pull/935)
* Add support getting the logs of container to cri manager [\#928](https://github.com/alibaba/pouch/pull/928)
* Add support setting pod dns configuration to cri manager [\#912](https://github.com/alibaba/pouch/pull/912)
* Wrap cri manager to log every cri operation [\#899](https://github.com/alibaba/pouch/pull/899)
* Fix inspect image by image id with prefix [\#895](https://github.com/alibaba/pouch/pull/895)
* Implement exec and attach method of stream server [\#854](https://github.com/alibaba/pouch/pull/854)
* Add `--group-add` flag to `pouch create` command and supplemental groups for cri manager [\#753](https://github.com/alibaba/pouch/pull/753)

### Test

* Add mock test for `rename` client [\#1021](https://github.com/alibaba/pouch/pull/1021)
* Add mock test for `version` client [\#1004](https://github.com/alibaba/pouch/pull/1004)
* Add test cases for `imageCache.get` [\#979](https://github.com/alibaba/pouch/pull/979)
* Add mock test for client package [\#965](https://github.com/alibaba/pouch/pull/965)
* Add test case for `login/logout` command [\#908](https://github.com/alibaba/pouch/pull/908)
* Add related functions for test pouch daemon [\#884](https://github.com/alibaba/pouch/pull/884)
* Print error log in CI for debug [\#883](https://github.com/alibaba/pouch/pull/883)

## 0.2.1 (2018-03-09)

### Network

* Support port mapping and exposed ports in container

### Bugfix

* Fix project quota can't be set on kernel-4.9
* Fix rich container mode can't find binary in PATH

## 0.2.0 (2018-03-02)

### Runtime

* Add rich container mode for daemon and runc
* Add support for Intel RDT isolation
* Support add annotation for oci-specs in daemon
* Add memory limit options specifically for open source AliOS
* Add user group support for container
* Add image pulling proxy for Dragonfly
* Add sccomp support for container
* refactor package reference image to cover more scenarios
* Add privileged mode support for container
* Add capability support for container
* Add apparmor support for container
* Add blkio isolation support for container
* Add sysctl support for container
* Add more fields in ImageInfo struct
* support user setting default registry
* Add ipc, pid, uts namespace support for container

### Client

* Add login/logout command for registry
* Add update command for container's resource or restart policy and so on
* Support context in client side
* Add volume list command

### Network

* support host/none/container network mode

### Storage

* support diskquota via project quota and group quota only for local volume (container diskquota is in progress)

### Kubernetes(CRI)

* Add CNI framework implementation
* Add all options of container in CRI manager
* Using cri-tools to verify every interface implementation of CRI

### Document

* Add document pouch with LXCFS
* Add document how to install Pouch in Kubernetes cluster
* Add volume design document
* Add document pouch with rich container

## 0.1.0 (2018-01-17)

Initial experiemental release for public

* Initial implemention to integrate containerd 1.0 in daemon
* Hypervisor-based container implementation
* Achieve container resource view isolation via supporting LXCFS
* Add API and CLI documentation
* Add unit test for project
* Add API and CLI for project
* Implement basic CRI to support Kubernetes
* Be consistent with Moby's 1.12.6 API
* Support basic network management and volume management
* Make Pouch run as a system service
* Make Pouch installed on distribution CentOS 7.2 and Ubuntu 16.04
