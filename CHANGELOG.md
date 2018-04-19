# CHANGELOG

## 0.4.0 (2018-04-19)

### Remote API && Client

* Add `lxcfs` enabled info to `info` command [\#1152](https://github.com/alibaba/pouch/pull/1152) ([zhuangqh](https://github.com/zhuangqh))
* Add `snapshotter` info to pouch `inspect` command [\#1130](https://github.com/alibaba/pouch/pull/1130) ([HusterWan](https://github.com/HusterWan))
* Add `--rm` flag to pouch `run` command [\#1125](https://github.com/alibaba/pouch/pull/1125) ([Ace-Tang](https://github.com/Ace-Tang))
* Fix make inspect output to an array [\#1119](https://github.com/alibaba/pouch/pull/1119) ([faycheng](https://github.com/faycheng))
* Add image proxy info to pouch `info` command [\#1102](https://github.com/alibaba/pouch/pull/1102) ([ZouRui89](https://github.com/ZouRui89))
* Add `--volumes-from` flags to pouch `create` command [\#1131](https://github.com/alibaba/pouch/pull/1131) ([rudyfly](https://github.com/rudyfly))
* Add `--cpu-period` and `--cpu-quota` to pouch `create` command [\#1067](https://github.com/alibaba/pouch/pull/1067) ([allencloud](https://github.com/allencloud))
* Refactor move parameters parse and validate part into `opts` package [\#1041](https://github.com/alibaba/pouch/pull/1041) ([HusterWan](https://github.com/HusterWan))
* Fix `image inspect` and `network inspect` command docs [\#1053](https://github.com/alibaba/pouch/pull/1053) ([HusterWan](https://github.com/HusterWan))
* Fix restful api url should support both with or without version info [\#1035](https://github.com/alibaba/pouch/pull/1035) ([HusterWan](https://github.com/HusterWan))
* Fix client login logic [\#1044](https://github.com/alibaba/pouch/pull/1044) ([Ace-Tang](https://github.com/Ace-Tang))
* Add `--annotation` to pouch `create` command [\#1046](https://github.com/alibaba/pouch/pull/1046) ([Ace-Tang](https://github.com/Ace-Tang))
* Fix make json ID be Id to be compatible with Moby API [\#1066](https://github.com/alibaba/pouch/pull/1066) ([allencloud](https://github.com/allencloud))
* Fix make pouch `create` output to an array [\#1034](https://github.com/alibaba/pouch/pull/1034) ([ZouRui89](https://github.com/ZouRui89))
* Add more daemon option in `info` API [\#1122](https://github.com/alibaba/pouch/pull/1122) ([allencloud](https://github.com/allencloud))
* Add more informations in `volume list` result [\#1028](https://github.com/alibaba/pouch/pull/1028) ([rudyfly](https://github.com/rudyfly))
* Fix modify `volume inspect` docs [\#1029](https://github.com/alibaba/pouch/pull/1029) ([rudyfly](https://github.com/rudyfly))

### Runtime

* Fix errors when using `volume-from` creates container[\#1161](https://github.com/alibaba/pouch/pull/1161) ([rudyfly](https://github.com/rudyfly))
* Fix set quota-id without disk-quota [\#1157](https://github.com/alibaba/pouch/pull/1157) ([rudyfly](https://github.com/rudyfly))
* Upgrade `containerd` vendor version to `v1.0.3` [\#1148](https://github.com/alibaba/pouch/pull/1148) ([fuweid](https://github.com/fuweid))
* Refactor move `pkg/opts` package to `api/opts` [\#1147](https://github.com/alibaba/pouch/pull/1147) ([allencloud](https://github.com/allencloud))
* Add support parsing volumes from docker image [\#1145](https://github.com/alibaba/pouch/pull/1145) ([rudyfly](https://github.com/rudyfly))
* Fix code style: we should not define a empty slice with `make` [\#1142](https://github.com/alibaba/pouch/pull/1142) ([oiooj](https://github.com/oiooj))
* Fix pouchd panic when meta snapshotter is nil [\#1140](https://github.com/alibaba/pouch/pull/1140) ([oiooj](https://github.com/oiooj))
* Fix set diskquota failed without quota id [\#1136](https://github.com/alibaba/pouch/pull/1136) ([rudyfly](https://github.com/rudyfly))
* Add node ip and sn into daemon labels [\#1134](https://github.com/alibaba/pouch/pull/1134) ([allencloud](https://github.com/allencloud))
* Refactor config file resolve [\#1132](https://github.com/alibaba/pouch/pull/1132) ([Ace-Tang](https://github.com/Ace-Tang))
* Add support to gc unused exec processes [\#1129](https://github.com/alibaba/pouch/pull/1129) ([Ace-Tang](https://github.com/Ace-Tang))
* Make TLS config params can be setup in the file [\#1126](https://github.com/alibaba/pouch/pull/1126) ([yyb196](https://github.com/yyb196))
* Add plugin point before endpoint creating [\#1124](https://github.com/alibaba/pouch/pull/1124) ([yyb196](https://github.com/yyb196))
* Fix return err when `ExecContainer` failed [\#1117](https://github.com/alibaba/pouch/pull/1117) ([oblivionfallout](https://github.com/oblivionfallout))
* Fix remove ip mask in `Networks.IPAddress` [\#1116](https://github.com/alibaba/pouch/pull/1116) ([rudyfly](https://github.com/rudyfly))
* Setup profiler and don't bother to enable debug level log [\#1111](https://github.com/alibaba/pouch/pull/1111) ([yyb196](https://github.com/yyb196))
* Fix we should do not append `latest` tag to the image when it already has a tag [\#1110](https://github.com/alibaba/pouch/pull/1110) ([yyb196](https://github.com/yyb196))
* Fix make container exit with real exit code [\#1099](https://github.com/alibaba/pouch/pull/1099) ([Ace-Tang](https://github.com/Ace-Tang))
* Add more flags in daemon config file [\#1088](https://github.com/alibaba/pouch/pull/1088) ([Ace-Tang](https://github.com/Ace-Tang))
* bugfix: fix interface casting error [\#1085](https://github.com/alibaba/pouch/pull/1085) ([Ace-Tang](https://github.com/Ace-Tang))
* Fix `gocyclo` issues in go report [\#1084](https://github.com/alibaba/pouch/pull/1084) ([zhuangqh](https://github.com/zhuangqh))
* Fix some spell errors [\#1081](https://github.com/alibaba/pouch/pull/1081) ([dbdd4us](https://github.com/dbdd4us))
* Add allinone to deploy pouch as available container to kubernetes [\#1070](https://github.com/alibaba/pouch/pull/1070) ([ZouRui89](https://github.com/ZouRui89))
* Fix golint not found has exit code 1 [\#1059](https://github.com/alibaba/pouch/pull/1059) ([sunyuan3](https://github.com/sunyuan3))
* Add `diskquota` support regular expression [\#1057](https://github.com/alibaba/pouch/pull/1057) ([rudyfly](https://github.com/rudyfly))
* Fix sync abnormal containers status when start pouchd [\#1056](https://github.com/alibaba/pouch/pull/1056) ([HusterWan](https://github.com/HusterWan))
* Remove duplicate error messages in network package [\#1048](https://github.com/alibaba/pouch/pull/1048) ([faycheng](https://github.com/faycheng))
* Fix setup hook in advance to avoid panic if no prestart hook setup before [\#1038](https://github.com/alibaba/pouch/pull/1038) ([yyb196](https://github.com/yyb196))
* Enable setup common name whitelist for tls checking [\#1036](https://github.com/alibaba/pouch/pull/1036) ([yyb196](https://github.com/yyb196))

### Network

* Refactor manage `libnetwork` by subtree instead of submodule [\#1135](https://github.com/alibaba/pouch/pull/1135) ([rudyfly](https://github.com/rudyfly))
* Fix make pouch network non-existent return exit code 1 [\#1089](https://github.com/alibaba/pouch/pull/1089) ([allencloud](https://github.com/allencloud))
* Fix delete endpoint after failing to create endpoint [\#1069](https://github.com/alibaba/pouch/pull/1069) ([faycheng](https://github.com/faycheng))
* Add support for inspecting network by ID [\#1040](https://github.com/alibaba/pouch/pull/1040) ([faycheng](https://github.com/faycheng))

### Kubernetes

* Fix make infra image configurable [\#1159](https://github.com/alibaba/pouch/pull/1159) ([YaoZengzeng](https://github.com/YaoZengzeng))
* Add `--enable-cri` flag to `pouchd` to specify whether enable CRI [\#1118](https://github.com/alibaba/pouch/pull/1118) ([YaoZengzeng](https://github.com/YaoZengzeng))
* Fix we should get image ID from `containerd` [\#1112](https://github.com/alibaba/pouch/pull/1112) ([YaoZengzeng](https://github.com/YaoZengzeng))
* Add image auth for cri manager [\#1097](https://github.com/alibaba/pouch/pull/1097) ([YaoZengzeng](https://github.com/YaoZengzeng))
* Add selinux support for cri manager [\#1092](https://github.com/alibaba/pouch/pull/1092) ([YaoZengzeng](https://github.com/YaoZengzeng))
* Fix loop `InspectExec` to finish in `ExecSync` and `Exec` operations [\#1086](https://github.com/alibaba/pouch/pull/1086) ([YaoZengzeng](https://github.com/YaoZengzeng))
* Fix for `privileged` container, make `dir` prefix with `/sys` ReadWrite [\#1055](https://github.com/alibaba/pouch/pull/1055) ([YaoZengzeng](https://github.com/YaoZengzeng))
* Fix distinguish `cmd` and `entrypoint` better [\#1045](https://github.com/alibaba/pouch/pull/1045) ([YaoZengzeng](https://github.com/YaoZengzeng))
* Fix specify both `cmd` and `args` when create a container [\#1027](https://github.com/alibaba/pouch/pull/1027) ([YaoZengzeng](https://github.com/YaoZengzeng))

### Test

* Add `lxcfs` binary check [\#1127](https://github.com/alibaba/pouch/pull/1127) ([Letty5411](https://github.com/Letty5411))
* Add `tls` test [\#1115](https://github.com/alibaba/pouch/pull/1115) ([Letty5411](https://github.com/Letty5411))
* Add mock test for `create` client [\#1106](https://github.com/alibaba/pouch/pull/1106) ([Dewey-Ding](https://github.com/Dewey-Ding))
* Seperate integration test and cri test in travisCI [\#1098](https://github.com/alibaba/pouch/pull/1098) ([Letty5411](https://github.com/Letty5411))
* Add mock test for `top` client [\#1093](https://github.com/alibaba/pouch/pull/1093) ([zhuangqh](https://github.com/zhuangqh))
* Add mock test for `resize` and `restart` client [\#1090](https://github.com/alibaba/pouch/pull/1090) ([Dewey-Ding](https://github.com/Dewey-Ding))
* Add tests for `label` and config file in `pouchd` [\#1087](https://github.com/alibaba/pouch/pull/1087) ([Letty5411](https://github.com/Letty5411))
* Refine `pouchd` test framework [\#1078](https://github.com/alibaba/pouch/pull/1078) ([Letty5411](https://github.com/Letty5411))
* Add mock test `pause` and `unpause` client [\#1074](https://github.com/alibaba/pouch/pull/1074) ([Dewey-Ding](https://github.com/Dewey-Ding))
* Add mock test for `remove` and `stop` client [\#1064](https://github.com/alibaba/pouch/pull/1064) ([Dewey-Ding](https://github.com/Dewey-Ding))
* Modify hardcode testcase for device `/dev/sda` [\#1054](https://github.com/alibaba/pouch/pull/1054) ([Ace-Tang](https://github.com/Ace-Tang))
* Add mock test for `list` client [\#1049](https://github.com/alibaba/pouch/pull/1049) ([Dewey-Ding](https://github.com/Dewey-Ding))
* Add mock test for `get` client [\#1037](https://github.com/alibaba/pouch/pull/1037) ([Dewey-Ding](https://github.com/Dewey-Ding))
* Add mock test for `image` operations on client side [\#1032](https://github.com/alibaba/pouch/pull/1032) ([ZouRui89](https://github.com/ZouRui89))
* Add mock test for `volume` operations on client side [\#1026](https://github.com/alibaba/pouch/pull/1026) ([ZouRui89](https://github.com/ZouRui89))
* Add mock test for `update` client [\#1024](https://github.com/alibaba/pouch/pull/1024) ([Dewey-Ding](https://github.com/Dewey-Ding))
* Add unit test in circleci [\#1039](https://github.com/alibaba/pouch/pull/1039) ([ZouRui89](https://github.com/ZouRui89))
* Add circleci parallel testing to split unit-test and code-check [\#1043](https://github.com/alibaba/pouch/pull/1043) ([allencloud](https://github.com/allencloud))
* Fix logic error in `volume create` mock test [\#1033](https://github.com/alibaba/pouch/pull/1033) ([ZouRui89](https://github.com/ZouRui89))
* Add restricts in codecov.yml to ignore files [\#1050](https://github.com/alibaba/pouch/pull/1050) ([allencloud](https://github.com/allencloud))

### New Contributors

Here is the list of new contributors:

* [Dewey-Ding](https://github.com/Dewey-Ding)
* [oiooj](https://github.com/oiooj)
* [dbdd4us](https://github.com/dbdd4us))
* [zhuangqh](https://github.com/zhuangqh)
* [oblivionfallout](https://github.com/oblivionfallout)

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
