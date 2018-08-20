# CHANGELOG

## 1.0.0-rc1 (2018-07-13)

__IMPORTANT__: In PouchContainer 1.0.0-rc1 we have done many things that important to all users of PouchContainer:

* PouchContainer-CRI now supports using annotation to choose different runtime.
* PouchContainer Image Manager supports `load` and `save` functionality.
* PouchContainer Log Driver supports `syslog` type.
* PouchContainer uses latest `libnetwork` in network module.
* PouchContainer makes the runtime module stable.

### __Pouch Daemon & API:__

* bugfix: Fix log-opt option parse fails if value contains comma [#1729](https://github.com/alibaba/pouch/pull/1729) ([Frankzhaopku](https://github.com/Frankzhaopku))
* enhance: add ExtraAttribute functionality in LogInfo [#1714](https://github.com/alibaba/pouch/pull/1714) ([fuweid](https://github.com/fuweid))
* bugfix: schema point to a response object  [#1712](https://github.com/alibaba/pouch/pull/1712) ([zhuangqh](https://github.com/zhuangqh))
* bugfix: fix exec record user as container config user [#1657](https://github.com/alibaba/pouch/pull/1657) ([Ace-Tang](https://github.com/Ace-Tang))
* feature: deamon support --log-driver and --log-opt options [#1647](https://github.com/alibaba/pouch/pull/1647) ([zhuangqh](https://github.com/zhuangqh))
* bugfix: list image should ignore error if containerd can't handle well [#1625](https://github.com/alibaba/pouch/pull/1625) ([fuweid](https://github.com/fuweid))
* bugfix: execConfig remove omitemtpy [#1619](https://github.com/alibaba/pouch/pull/1619) ([HusterWan](https://github.com/HusterWan))
* enhance: add new formatter for syslog [#1608](https://github.com/alibaba/pouch/pull/1608) ([fuweid](https://github.com/fuweid))
* enhance: Port Pouch Cli to Darwin(MacOS) [#1598](https://github.com/alibaba/pouch/pull/1598) ([xuzhenglun](https://github.com/xuzhenglun))
* feature: support pouch ps filter [#1595](https://github.com/alibaba/pouch/pull/1595) ([Ace-Tang](https://github.com/Ace-Tang))
* feature: add pouch save functionality [#1592](https://github.com/alibaba/pouch/pull/1592) ([xiechengsheng](https://github.com/xiechengsheng))
* enhance: adjust data stream from pouch pull api [#1586](https://github.com/alibaba/pouch/pull/1586) ([fuweid](https://github.com/fuweid))
* feature: add systemd notify [#1577](https://github.com/alibaba/pouch/pull/1577) ([shaloulcy](https://github.com/shaloulcy))
* feature: adjust pouchd unix socket permissions [#1561](https://github.com/alibaba/pouch/pull/1561) ([Ace-Tang](https://github.com/Ace-Tang))
* bugfix: rename cpu-share to cpu-shares in command line [#1547](https://github.com/alibaba/pouch/pull/1547) ([allencloud](https://github.com/allencloud))
* feature: add update daemon config function [#1514](https://github.com/alibaba/pouch/pull/1514) ([rudyfly](https://github.com/rudyfly))
* bugfix: pull image before run and upgrade [#1419](https://github.com/alibaba/pouch/pull/1419) ([wrfly](https://github.com/wrfly))
* feature: add pouch load functionality [#1391](https://github.com/alibaba/pouch/pull/1391) ([fuweid](https://github.com/fuweid))
* feature: add wait client command for pouch [#1333](https://github.com/alibaba/pouch/pull/1333) ([xiechengsheng](https://github.com/xiechengsheng))

### __Container Runtime:__

* bugfix: execute remount rootfs before prestart hook [#1622](https://github.com/alibaba/pouch/pull/1622) ([HusterWan](https://github.com/HusterWan))
* bugfix: release the container resources if contaienr failed to start [#1621](https://github.com/alibaba/pouch/pull/1621) ([shaloulcy](https://github.com/shaloulcy))
* bugfix: fix memory-swap flag not validate correct [#1614](https://github.com/alibaba/pouch/pull/1614) ([Ace-Tang](https://github.com/Ace-Tang))
* bugfix: fix exec stuck when exec get error [#1605](https://github.com/alibaba/pouch/pull/1605) ([Ace-Tang](https://github.com/Ace-Tang))
* bugfix: we should set Running flag to true when started container [#1604](https://github.com/alibaba/pouch/pull/1604) ([HusterWan](https://github.com/HusterWan))
* refactor: move config file from cli into one places [#1597](https://github.com/alibaba/pouch/pull/1597) ([Ace-Tang](https://github.com/Ace-Tang))
* feature: support net priority flag [#1576](https://github.com/alibaba/pouch/pull/1576) ([Ace-Tang](https://github.com/Ace-Tang))
* refactor: refactor container update diskquota type [#1572](https://github.com/alibaba/pouch/pull/1572) ([rudyfly](https://github.com/rudyfly))
* bugfix: make ringbuffer right [#1558](https://github.com/alibaba/pouch/pull/1558) ([fuweid](https://github.com/fuweid))
* bugfix: vendor latest libnetwork for connect panic [#1556](https://github.com/alibaba/pouch/pull/1556) ([shaloulcy](https://github.com/shaloulcy))
* feature: support shm size [#1542](https://github.com/alibaba/pouch/pull/1542) ([Ace-Tang](https://github.com/Ace-Tang))
* bugfix: change the order of generating MountPoints [#1541](https://github.com/alibaba/pouch/pull/1541) ([shaloulcy](https://github.com/shaloulcy))
* refactor: make code more encapsulate and logic simple [#1540](https://github.com/alibaba/pouch/pull/1540) ([allencloud](https://github.com/allencloud))
* feature: add container setting check [#1537](https://github.com/alibaba/pouch/pull/1537) ([Ace-Tang](https://github.com/Ace-Tang))
* bugfix: update cpu-quota of 0.2.4 container may occur error [#1533](https://github.com/alibaba/pouch/pull/1533) ([HusterWan](https://github.com/HusterWan))
* enhance: add volume lock [#1531](https://github.com/alibaba/pouch/pull/1531) ([shaloulcy](https://github.com/shaloulcy))
* bugfix: restore config after update fail [#1513](https://github.com/alibaba/pouch/pull/1513) ([Ace-Tang](https://github.com/Ace-Tang))
* bugfix: rm omitempty in resource fields [#1505](https://github.com/alibaba/pouch/pull/1505) ([allencloud](https://github.com/allencloud))
* bugfix: support nullable bool value set in config [#1502](https://github.com/alibaba/pouch/pull/1502) ([Ace-Tang](https://github.com/Ace-Tang))
* feature: init syslog functionality in pouchd [#1500](https://github.com/alibaba/pouch/pull/1500) ([fuweid](https://github.com/fuweid))
* bugfix: fix the wrong bridge gateway [#1495](https://github.com/alibaba/pouch/pull/1495) ([rudyfly](https://github.com/rudyfly))
* bugfix: set memory swap double of non-zero memory value [#1492](https://github.com/alibaba/pouch/pull/1492) ([Ace-Tang](https://github.com/Ace-Tang))
* bugfix: add attach volume when container start [#1483](https://github.com/alibaba/pouch/pull/1483) ([rudyfly](https://github.com/rudyfly))
* feature: support creating container by just specifying rootfs [#1474](https://github.com/alibaba/pouch/pull/1474) ([HusterWan](https://github.com/HusterWan))
* feature: finish the CLI logs part [#1472](https://github.com/alibaba/pouch/pull/1472) ([fuweid](https://github.com/fuweid))
* bugfix: copy data before put it into ringbuf [#1471](https://github.com/alibaba/pouch/pull/1471) ([fuweid](https://github.com/fuweid))
* bugfix: set memory swap initial value to 0 [#1466](https://github.com/alibaba/pouch/pull/1466) ([Ace-Tang](https://github.com/Ace-Tang))
* refactor: remove ceph volume plugin [#1441](https://github.com/alibaba/pouch/pull/1441) ([shaloulcy](https://github.com/shaloulcy))
* bugfix: we must call Restore container after initialize network Mgr [#1422](https://github.com/alibaba/pouch/pull/1422) ([HusterWan](https://github.com/HusterWan))
* bugfix: change the option to set volume size [#1409](https://github.com/alibaba/pouch/pull/1409) ([rudyfly](https://github.com/rudyfly))
* feature: add cgroup resources check [#1375](https://github.com/alibaba/pouch/pull/1375) ([Ace-Tang](https://github.com/Ace-Tang))
* feature: add runtime config [#1366](https://github.com/alibaba/pouch/pull/1366) ([Ace-Tang](https://github.com/Ace-Tang))
* feature: update lock policy for container management [#1307](https://github.com/alibaba/pouch/pull/1307) ([allencloud](https://github.com/allencloud))

### __Network:__

* enhance: display disconnect result for pouch network disconnect [#1590](https://github.com/alibaba/pouch/pull/1590) ([shaloulcy](https://github.com/shaloulcy))
* bugfix: network not found [#1473](https://github.com/alibaba/pouch/pull/1473) ([shaloulcy](https://github.com/shaloulcy))
* refactor: use govendor to depend on libnetwork [#1445](https://github.com/alibaba/pouch/pull/1445) ([idealhack](https://github.com/idealhack))
* feature: modify defaut bridge mode [#1424](https://github.com/alibaba/pouch/pull/1424) ([rudyfly](https://github.com/rudyfly))
* feature: add container's network files [#1403](https://github.com/alibaba/pouch/pull/1403) ([shaloulcy](https://github.com/shaloulcy))

### __Kubernetes CRI:__

* feature: make runtime choosing supported in CRI managers for Kubernetes [#1593](https://github.com/alibaba/pouch/pull/1593) ([Starnop](https://github.com/Starnop))
* bugfix: skip teardown network, if the sandbox has been stopped [#1539](https://github.com/alibaba/pouch/pull/1539) ([YaoZengzeng](https://github.com/YaoZengzeng))
* refactor: return CRI services error [#1521](https://github.com/alibaba/pouch/pull/1521) ([oiooj](https://github.com/oiooj))
* bugfix: bind an address for stream server [#1520](https://github.com/alibaba/pouch/pull/1520) ([YaoZengzeng](https://github.com/YaoZengzeng))
* feature: UpdateContainerResources of CRI Manager [#1511](https://github.com/alibaba/pouch/pull/1511) ([Starnop](https://github.com/Starnop))
* bugfix: Make hack/kubernetes/allinone\_aliyun.sh to pass shellcheck [#1507](https://github.com/alibaba/pouch/pull/1507) ([Starnop](https://github.com/Starnop))
* bugfix: if run sandbox failed, clean up; deduplicate the default mounts with user defined ones [#1468](https://github.com/alibaba/pouch/pull/1468) ([YaoZengzeng](https://github.com/YaoZengzeng))
* feature: stats of cri manager [#1431](https://github.com/alibaba/pouch/pull/1431) ([Starnop](https://github.com/Starnop))

### __Test & Tool:__

* bugfix: deb package build failed [#1727](https://github.com/alibaba/pouch/pull/1727) ([shaloulcy](https://github.com/shaloulcy))
* enhance: add .DS\_Store ignore [#1724](https://github.com/alibaba/pouch/pull/1724) ([Frankzhaopku](https://github.com/Frankzhaopku))
* test: add unit test for filter validation [#1718](https://github.com/alibaba/pouch/pull/1718) ([allencloud](https://github.com/allencloud))
* test: TestListVolumes [#1707](https://github.com/alibaba/pouch/pull/1707) ([mengjiahao](https://github.com/mengjiahao))
* test: add unit-test for ValidateCPUQuota [#1692](https://github.com/alibaba/pouch/pull/1692) ([johanzhu](https://github.com/johanzhu))
* Add unit-test for ValidateCPUPeriod [#1690](https://github.com/alibaba/pouch/pull/1690) ([johanzhu](https://github.com/johanzhu))
* test: TestRemoveVolume [#1679](https://github.com/alibaba/pouch/pull/1679) ([mengjiahao](https://github.com/mengjiahao))
* test: add unit-test for proxy/has port [#1668](https://github.com/alibaba/pouch/pull/1668) ([mengjiahao](https://github.com/mengjiahao))
* test: add unit-tests for core GetVolume [#1660](https://github.com/alibaba/pouch/pull/1660) ([forienlauo](https://github.com/forienlauo))
* test: add unit test for function ValidateOOMScore in oom\_score\_test.go [#1658](https://github.com/alibaba/pouch/pull/1658) ([quyi1993](https://github.com/quyi1993))
* test: add unit-test for hasPort method [#1654](https://github.com/alibaba/pouch/pull/1654) ([forienlauo](https://github.com/forienlauo))
* test: add unit-tests for core CreateVolume [#1626](https://github.com/alibaba/pouch/pull/1626) ([shaloulcy](https://github.com/shaloulcy))
* bugfix: should remove the container in specified daemon [#1613](https://github.com/alibaba/pouch/pull/1613) ([Letty5411](https://github.com/Letty5411))
* test: add top command in upgrade test suite [#1554](https://github.com/alibaba/pouch/pull/1554) ([allencloud](https://github.com/allencloud))
* test: add test for different volume sources [#1553](https://github.com/alibaba/pouch/pull/1553) ([shaloulcy](https://github.com/shaloulcy))
* test: add TestRunMemoryOOM test case [#1552](https://github.com/alibaba/pouch/pull/1552) ([sunyuan3](https://github.com/sunyuan3))
* test: add all states container restart validation [#1549](https://github.com/alibaba/pouch/pull/1549) ([allencloud](https://github.com/allencloud))
* bugfix: refine tests with specifying CMD [#1548](https://github.com/alibaba/pouch/pull/1548) ([Letty5411](https://github.com/Letty5411))
* bugfix: fix rpm package bug [#1519](https://github.com/alibaba/pouch/pull/1519) ([Letty5411](https://github.com/Letty5411))
* bugfix: rename lxcfs to pouch-lxcfs in pouch.rpm [#1490](https://github.com/alibaba/pouch/pull/1490) ([Letty5411](https://github.com/Letty5411))
* feature: add make help into makefile [#1478](https://github.com/alibaba/pouch/pull/1478) ([houstar](https://github.com/houstar))
* bugfix: replace DelContainerForceOk with DelContainerForceMultyTime [#1462](https://github.com/alibaba/pouch/pull/1462) ([zhuangqh](https://github.com/zhuangqh))
* test: add more test for container operations [#1457](https://github.com/alibaba/pouch/pull/1457) ([ZouRui89](https://github.com/ZouRui89))
* refactor: correct hard coding in several shell script [#1452](https://github.com/alibaba/pouch/pull/1452) ([zhuangqh](https://github.com/zhuangqh))
* bugfix: correct shell script format via shellcheck reports [#1447](https://github.com/alibaba/pouch/pull/1447) ([zhuangqh](https://github.com/zhuangqh))
* test: sort image list before check [#1413](https://github.com/alibaba/pouch/pull/1413) ([fuweid](https://github.com/fuweid))
* feature: travis doesn't run document-only changed commit [#1412](https://github.com/alibaba/pouch/pull/1412) ([fuweid](https://github.com/fuweid))

### __Documentation:__

* docs: change logos to new version [#1720](https://github.com/alibaba/pouch/pull/1720) ([Frankzhaopku](https://github.com/Frankzhaopku))
* Update [INSTALLATION.md](http://INSTALLATION.md) [#1698](https://github.com/alibaba/pouch/pull/1698) ([wq2526](https://github.com/wq2526))
* docs: update docs about contributing [#1673](https://github.com/alibaba/pouch/pull/1673) ([shannonxn](https://github.com/shannonxn))
* docs: enable non-root user to run pouch commands without sudo [#1573](https://github.com/alibaba/pouch/pull/1573) ([Ace-Tang](https://github.com/Ace-Tang))
* docs: change pouch to PouchContainer [#1525](https://github.com/alibaba/pouch/pull/1525) ([Frankzhaopku](https://github.com/Frankzhaopku))
* bugfix typo in the vendor/README.md [#1496](https://github.com/alibaba/pouch/pull/1496) ([houstar](https://github.com/houstar))
* docs: better arch to add connection between cri manager and pouchd [#1465](https://github.com/alibaba/pouch/pull/1465) ([allencloud](https://github.com/allencloud))
* docs: add docs about  lxcfs feature [#1461](https://github.com/alibaba/pouch/pull/1461) ([fanux](https://github.com/fanux))
* doc: Modify the document about Kubernetes&pouch to make it friendly [#1459](https://github.com/alibaba/pouch/pull/1459) ([Starnop](https://github.com/Starnop))
* docs: update [FAQ.md](http://FAQ.md) to add kernel version support [#1444](https://github.com/alibaba/pouch/pull/1444) ([allencloud](https://github.com/allencloud))
* docs: add supporting legacy kernels into runV [#1442](https://github.com/alibaba/pouch/pull/1442) ([allencloud](https://github.com/allencloud))
* docs: add more details on rich container [#1440](https://github.com/alibaba/pouch/pull/1440) ([allencloud](https://github.com/allencloud))

### New Contributors

Here is the list of new contributors:

* [Frankzhaopku](https://github.com/Frankzhaopku)
* [johanzhu](https://github.com/johanzhu)
* [lauo](https://github.com/forienlauo)
* [mengjiahao](https://github.com/mengjiahao)
* [xuzhenglun](https://github.com/xuzhenglun)
* [fanux](https://github.com/fanux)
* [wq2526](https://github.com/wq2526)
* [idealhack](https://github.com/idealhack)
* [shannonxn](https://github.com/shannonxn)

## 0.5.0 (2018-05-25)

**IMPORTANT**: In PouchContainer 0.5.0 we have done many things that important to all users of PouchContainer:

1. PouchContainer now supports CRI v1alpha2 that will support for Kubernetes 1.10.0
2. Add plugin mechanism that we can use many existing volume and network plugins
3. Add many container and image tools like `logs` and `tag` command that will be very helpful for daily container operation
4. PouchContainer now is more stable and works well in production environment

### Remote API && Client

* Add instruction comment for the `blkio-weight-device` flag of `run` command [\#1381](https://github.com/alibaba/pouch/pull/1381) ([Ace-Tang](https://github.com/Ace-Tang))
* Fix `cgroup-parent` can not be set from the daemon config file [\#1361](https://github.com/alibaba/pouch/pull/1361) ([Ace-Tang](https://github.com/Ace-Tang))
* Add volume drivers info to system info [\#1352](https://github.com/alibaba/pouch/pull/1352) ([shaloulcy](https://github.com/shaloulcy))
* Fix `ExecIDs` parameter of `ContainerConfig` should be a slice [\#1324](https://github.com/alibaba/pouch/pull/1324) ([HusterWan](https://github.com/HusterWan))
* Refactor format `topExamples` code and add `execExample` [\#1319](https://github.com/alibaba/pouch/pull/1319) ([soarpenguin](https://github.com/soarpenguin))
* Enhance add more field in `pouch info` command [\#1238](https://github.com/alibaba/pouch/pull/1238) ([ZouRui89](https://github.com/ZouRui89))
* Add `debug` flag to pouch client [\#1234](https://github.com/alibaba/pouch/pull/1234) ([shaloulcy](https://github.com/shaloulcy))
* Fix `exec` command align with Moby `v1.24` API [\#1226](https://github.com/alibaba/pouch/pull/1226) ([fuweid](https://github.com/fuweid))
* Fix volume info of `inspect` output is incompatible with Moby API [\#1213](https://github.com/alibaba/pouch/pull/1213) ([HusterWan](https://github.com/HusterWan))
* Refactor network list api, make it compatible with Mody API [\#1173](https://github.com/alibaba/pouch/pull/1173) ([rudyfly](https://github.com/rudyfly))

### Runtime

* Fix container may be killed when ontainerd instance exit [\#1407](https://github.com/alibaba/pouch/pull/1407) ([HusterWan](https://github.com/HusterWan))
* Fix panic when execute `exec` command with flag `-d` [\#1394](https://github.com/alibaba/pouch/pull/1394) ([HusterWan](https://github.com/HusterWan))
* New `tag` tool for pouch that allow create alias name for images [\#1378](https://github.com/alibaba/pouch/pull/1378) ([fuweid](https://github.com/fuweid))
* Fix add judge for whether pidfile path is given when start pouch daemon [\#1374](https://github.com/alibaba/pouch/pull/1374) ([Ace-Tang](https://github.com/Ace-Tang))
* Fix map type can not be merged [\#1367](https://github.com/alibaba/pouch/pull/1367) ([Ace-Tang](https://github.com/Ace-Tang))
* Add support for updating or deleting an env value [\#1364](https://github.com/alibaba/pouch/pull/1364) ([HusterWan](https://github.com/HusterWan))
* Add support for managing more containers in some commands [\#1357](https://github.com/alibaba/pouch/pull/1357) ([xiechengsheng](https://github.com/xiechengsheng))
* Fix remove `pids-limit` initial value [\#1354](https://github.com/alibaba/pouch/pull/1354) ([Ace-Tang](https://github.com/Ace-Tang))
* Add support for using an image by digest id [\#1351](https://github.com/alibaba/pouch/pull/1351) ([fuweid](https://github.com/fuweid))
* Support generate version information at build time [\#1350](https://github.com/alibaba/pouch/pull/1350) ([Ace-Tang](https://github.com/Ace-Tang))
* Fix let `execConfig` value assignment before IO close, so thath `CRI` test case can get right result before container exit [\#1340](https://github.com/alibaba/pouch/pull/1340) ([ZouRui89](https://github.com/ZouRui89))
* Fix make the jsonfile exit friendly [\#1330](https://github.com/alibaba/pouch/pull/1330) ([fuweid](https://github.com/fuweid))
* Fix `markStopped` may failed that will cause container status not right [\#1322](https://github.com/alibaba/pouch/pull/1322) ([HusterWan](https://github.com/HusterWan))
* Fix mount `/sys/fs/cgroup` into container [\#1314](https://github.com/alibaba/pouch/pull/1314) ([fuweid](https://github.com/fuweid))
* Fix make pouch daemon exit friendly [\#1311](https://github.com/alibaba/pouch/pull/1311) ([fuweid](https://github.com/fuweid))
* Refactor eliminate `containerMeta` in pouch daemon manager [\#1300](https://github.com/alibaba/pouch/pull/1300) ([allencloud](https://github.com/allencloud))
* New `logs` API implement to redirct container's `StdOut` and `StdErr` to json file [\#1298](https://github.com/alibaba/pouch/pull/1298) ([fuweid](https://github.com/fuweid))
* Refactor reorder function sequence to make it more reasonable [\#1296](https://github.com/alibaba/pouch/pull/1296) ([allencloud](https://github.com/allencloud))
* Add support for taking over old containerd instance when pouchd restart [\#1275](https://github.com/alibaba/pouch/pull/1275) ([HusterWan](https://github.com/HusterWan))
* Fix can't stop a `paused` container [\#1269](https://github.com/alibaba/pouch/pull/1269) ([shaloulcy](https://github.com/shaloulcy))
* Refactor image manager: redesign the `imageStore` in image manager and make it more clear and stable [\#1267](https://github.com/alibaba/pouch/pull/1267) ([fuweid](https://github.com/fuweid))
* Fix compatibility with alidocker when update container diskquota [\#1264](https://github.com/alibaba/pouch/pull/1264) ([HusterWan](https://github.com/HusterWan))
* Fix change the `QuotaID` to `QuotaId` to align with Moby `v1.24` API [\#1263](https://github.com/alibaba/pouch/pull/1263) ([fuweid](https://github.com/fuweid))
* Refactor facilitate make.sh `build` part code [\#1261](https://github.com/alibaba/pouch/pull/1261) ([u2takey](https://github.com/u2takey))
* Fix set container env failed because `Invalid cross-device link` error [\#1260](https://github.com/alibaba/pouch/pull/1260) ([HusterWan](https://github.com/HusterWan))
* Fix the output file is used incorrectly & fix the wrong test case name [\#1258](https://github.com/alibaba/pouch/pull/1258) ([xieyanke](https://github.com/xieyanke))
* Add `--volume` flag when remove container that will delete all anonymous volumes created by pouchd [\#1255](https://github.com/alibaba/pouch/pull/1255) ([rudyfly](https://github.com/rudyfly))
* Fix mountpoint binary not found error [\#1253](https://github.com/alibaba/pouch/pull/1253) ([shaloulcy](https://github.com/shaloulcy))
* Fix the mount path of tmpfs volume and some misspells [\#1251](https://github.com/alibaba/pouch/pull/1251) ([shaloulcy](https://github.com/shaloulcy))
* Fix merge flag default value in pouch daemon config, if flag not be passed, we should not merge it with daemon config [\#1246](https://github.com/alibaba/pouch/pull/1246) ([Ace-Tang](https://github.com/Ace-Tang))
* Add vagrant environment for development [\#1245](https://github.com/alibaba/pouch/pull/1245) ([u2takey](https://github.com/u2takey))
* Fix free resources after exec exit [\#1240](https://github.com/alibaba/pouch/pull/1240) ([Ace-Tang](https://github.com/Ace-Tang))
* Fix modify log format [\#1239](https://github.com/alibaba/pouch/pull/1239) ([rudyfly](https://github.com/rudyfly))
* Fix add newline for id when create container [\#1237](https://github.com/alibaba/pouch/pull/1237) ([fuweid](https://github.com/fuweid))
* Add update restful api support to update container diskquota [\#1235](https://github.com/alibaba/pouch/pull/1235) ([HusterWan](https://github.com/HusterWan))
* Fix `setRawMode` can only be set when the user set tty [\#1233](https://github.com/alibaba/pouch/pull/1233) ([fuweid](https://github.com/fuweid))
* Fix compatibility with `alidocker` when update container labels [\#1228](https://github.com/alibaba/pouch/pull/1228) ([HusterWan](https://github.com/HusterWan))
* Add `--pids-limit` flags to `create` command [\#1227](https://github.com/alibaba/pouch/pull/1227) ([Ace-Tang](https://github.com/Ace-Tang))
* Add support update container's `cpu-period` [\#1222](https://github.com/alibaba/pouch/pull/1222) ([HusterWan](https://github.com/HusterWan))
* Add support parsing `ContainerConfig.Volumes` when create container [\#1219](https://github.com/alibaba/pouch/pull/1219) ([rudyfly](https://github.com/rudyfly))
* Add support updating env when container is running [\#1218](https://github.com/alibaba/pouch/pull/1218) ([HusterWan](https://github.com/HusterWan))
* Fix volume size without unit [\#1215](https://github.com/alibaba/pouch/pull/1215) ([rudyfly](https://github.com/rudyfly))
* Fix remove the `DiskQuota` field from `Resource` struct [\#1212](https://github.com/alibaba/pouch/pull/1212) ([fuweid](https://github.com/fuweid))
* Fix make stopping an non-running container valid [\#1210](https://github.com/alibaba/pouch/pull/1210) ([allencloud](https://github.com/allencloud))
* Fix make stopping a stopped container return no error [\#1209](https://github.com/alibaba/pouch/pull/1209) ([allencloud](https://github.com/allencloud))
* Fix make `restart` API support restarting an stopped container [\#1208](https://github.com/alibaba/pouch/pull/1208) ([allencloud](https://github.com/allencloud))
* Refactor store container info to disk failed should return errors [\#1203](https://github.com/alibaba/pouch/pull/1203) ([HusterWan](https://github.com/HusterWan))
* Fix make logrus detect whether output with color [\#1202](https://github.com/alibaba/pouch/pull/1202) ([yyb196](https://github.com/yyb196))
* Fix remove update image and fix bugs when update env [\#1196](https://github.com/alibaba/pouch/pull/1196) ([HusterWan](https://github.com/HusterWan))
* Fix remove log format type check [\#1192](https://github.com/alibaba/pouch/pull/1192) ([oiooj](https://github.com/oiooj))
* Fix add default tag `:latest` when using pouch `rmi` command to remove untagged container images. [\#1191](https://github.com/alibaba/pouch/pull/1191) ([xiechengsheng](https://github.com/xiechengsheng))
* Fix container cannot start after first start failed [\#1190](https://github.com/alibaba/pouch/pull/1190) ([HusterWan](https://github.com/HusterWan))
* Fix check the duplicate mount point [\#1185](https://github.com/alibaba/pouch/pull/1185) ([rudyfly](https://github.com/rudyfly))
* Fix if user rename container with id use the real name to clean the cache [\#1182](https://github.com/alibaba/pouch/pull/1182) ([yyb196](https://github.com/yyb196))
* Add `--ulimit` flag to `create` command [\#1179](https://github.com/alibaba/pouch/pull/1179) ([Ace-Tang](https://github.com/Ace-Tang))
* Add support recording container's last exit time [\#1176](https://github.com/alibaba/pouch/pull/1176) ([Ace-Tang](https://github.com/Ace-Tang))
* Fix update `SafeMap` item should just call `Put` method, no need call `Remove` method [\#1175](https://github.com/alibaba/pouch/pull/1175) ([HusterWan](https://github.com/HusterWan))
* Add support setting volumes to `/etc/mtab` [\#1170](https://github.com/alibaba/pouch/pull/1170) ([rudyfly](https://github.com/rudyfly))

### Documentation

* Update doc of pouch with kubernetes deploying [\#1384](https://github.com/alibaba/pouch/pull/1384) ([Starnop](https://github.com/Starnop))
* Update `apt-key` fingerprint to `BE2F475F` when install pouch on ubuntu [\#1339](https://github.com/alibaba/pouch/pull/1339) ([rhinoceros](https://github.com/rhinoceros))
* Add more information about how to run test [\#1331](https://github.com/alibaba/pouch/pull/1331) ([Letty5411](https://github.com/Letty5411))
* Add misspell tool to check English words [\#1304](https://github.com/alibaba/pouch/pull/1304) ([allencloud](https://github.com/allencloud))
* Add introduction document about how to run `kata-container` with pouch [\#1295](https://github.com/alibaba/pouch/pull/1295) ([Ace-Tang](https://github.com/Ace-Tang))
* Add code style introduction document for pouch [\#1283](https://github.com/alibaba/pouch/pull/1283) ([allencloud](https://github.com/allencloud))
* Add introduction document about how to  deploy kubernetes with pouch powerd by aliyun [\#1236](https://github.com/alibaba/pouch/pull/1236) ([Starnop](https://github.com/Starnop))
* Fix typos [\#1177](https://github.com/alibaba/pouch/pull/1177) ([wgliang](https://github.com/wgliang))  [\#1178](https://github.com/alibaba/pouch/pull/1178) ([XSAM](https://github.com/XSAM))  [\#1200](https://github.com/alibaba/pouch/pull/1200) ([shaloulcy](https://github.com/shaloulcy))  [\#1189](https://github.com/alibaba/pouch/pull/1189) ([xiechengsheng](https://github.com/xiechengsheng))  [\#1303](https://github.com/alibaba/pouch/pull/1303) ([chuanchang](https://github.com/chuanchang))  [\#1248](https://github.com/alibaba/pouch/pull/1248) ([raoqi](https://github.com/raoqi))  [\#1216](https://github.com/alibaba/pouch/pull/1216) ([shaloulcy](https://github.com/shaloulcy))
* Fix pouch github address url  [\#1229](https://github.com/alibaba/pouch/pull/1229) ([u2takey](https://github.com/u2takey))
* Add markdownlint tool in Dockerfile [\#1204](https://github.com/alibaba/pouch/pull/1204) ([allencloud](https://github.com/allencloud))

### Kubernetes

* Fix return container `LogPath` in `ContainerStatusResponse` [\#1401](https://github.com/alibaba/pouch/pull/1401) ([Starnop](https://github.com/Starnop))
* Fix replace pod default `pause` image with the google released image [\#1382](https://github.com/alibaba/pouch/pull/1382) ([ZouRui89](https://github.com/ZouRui89))
* Add support both for CRI v1alpha1 and v1alpha2 version [\#1359](https://github.com/alibaba/pouch/pull/1359) ([Starnop](https://github.com/Starnop))
* Add timeout handler for `execSync` in cri part [\#1318](https://github.com/alibaba/pouch/pull/1318) ([ZouRui89](https://github.com/ZouRui89))
* Refactor move the `CRI` code out of pouch `mgr` dirctory [\#1317](https://github.com/alibaba/pouch/pull/1317) ([Starnop](https://github.com/Starnop))
* Fix disable mux stdout and stderr if backend is not http [\#1250](https://github.com/alibaba/pouch/pull/1250) ([YaoZengzeng](https://github.com/YaoZengzeng))
* Fix handle container io properly when restart pouchd [\#1220](https://github.com/alibaba/pouch/pull/1220) ([YaoZengzeng](https://github.com/YaoZengzeng))
* Fix evaluate finish time of container in CRI correctly [\#1183](https://github.com/alibaba/pouch/pull/1183) ([YaoZengzeng](https://github.com/YaoZengzeng))

### Storage

* Support volume plugin mechanism, now pouch not only supports `local/tmpfs/ceph` volumes, but also support existing mature docker volume drivers [\#1326](https://github.com/alibaba/pouch/pull/1326) ([shaloulcy](https://github.com/shaloulcy))
* Fix we should add lock before visit volume boltdb [\#1286](https://github.com/alibaba/pouch/pull/1286) ([rudyfly](https://github.com/rudyfly))
* Add pouch `plugin` mechanism, so that we can use existing moby volume and network plugins [\#1278](https://github.com/alibaba/pouch/pull/1278) ([shaloulcy](https://github.com/shaloulcy))
* Fix change volume metadata struct for remote storage manager [\#1271](https://github.com/alibaba/pouch/pull/1271) ([rudyfly](https://github.com/rudyfly))
* Add `volume-driver-alias` flag to volume manager, we can set alias name for the exist volume drivers [\#1224](https://github.com/alibaba/pouch/pull/1224) ([rudyfly](https://github.com/rudyfly))

### Network

* Fix endpoints are disappeared when pouchd restart [\#1312](https://github.com/alibaba/pouch/pull/1312) ([rudyfly](https://github.com/rudyfly))
* Fix remove all endpoints when execute `network disconnect` command [\#1284](https://github.com/alibaba/pouch/pull/1284) ([rudyfly](https://github.com/rudyfly))
* Add `network connect` interface for container [\#1187](https://github.com/alibaba/pouch/pull/1187) ([rudyfly](https://github.com/rudyfly))
* Add `network disconnect` interface for container [\#1172](https://github.com/alibaba/pouch/pull/1172) ([HusterWan](https://github.com/HusterWan))

### Test

* Fix use `busybox:1.25` instead of `busybox:1.28` in `tag` command CLI test [\#1406](https://github.com/alibaba/pouch/pull/1406) ([fuweid](https://github.com/fuweid))
* Fix use the stable image ID in test case [\#1397](https://github.com/alibaba/pouch/pull/1397) ([fuweid](https://github.com/fuweid))
* Fix make the PullImage test util work [\#1386](https://github.com/alibaba/pouch/pull/1386) ([fuweid](https://github.com/fuweid))
* Update split `run` command test file into several files [\#1385](https://github.com/alibaba/pouch/pull/1385) ([Letty5411](https://github.com/Letty5411))
* Add test cases for `volume plugin` [\#1368](https://github.com/alibaba/pouch/pull/1368) ([shaloulcy](https://github.com/shaloulcy))
* Add CLI test for `pause` command and fix some tiny bugs [\#1360](https://github.com/alibaba/pouch/pull/1360) ([ZouRui89](https://github.com/ZouRui89))
* Fix `TestRunWithPidsLimit` test case failed because no pids cgroup support [\#1353](https://github.com/alibaba/pouch/pull/1353) ([Ace-Tang](https://github.com/Ace-Tang))
* Enhance CLI related tests [\#1341](https://github.com/alibaba/pouch/pull/1341) ([Letty5411](https://github.com/Letty5411))
* Fix `ps` command CLI tests failed [\#1334](https://github.com/alibaba/pouch/pull/1334) ([HusterWan](https://github.com/HusterWan))
* Fix missing removal of container when test suit end [\#1327](https://github.com/alibaba/pouch/pull/1327) ([allencloud](https://github.com/allencloud))
* Fix using existing image and fix shell format error [\#1313](https://github.com/alibaba/pouch/pull/1313) ([Letty5411](https://github.com/Letty5411))
* Add `-race` flag to `go test` command to detect race [\#1294](https://github.com/alibaba/pouch/pull/1294) ([allencloud](https://github.com/allencloud))
* Enhance making test more robust [\#1279](https://github.com/alibaba/pouch/pull/1279) ([Letty5411](https://github.com/Letty5411))
* Fix `restart` the `paused` status container ci failed  [\#1272](https://github.com/alibaba/pouch/pull/1272) ([shaloulcy](https://github.com/shaloulcy))
* Fix `run` container exit because of no using long run command caused ci failed [\#1214](https://github.com/alibaba/pouch/pull/1214) ([Ace-Tang](https://github.com/Ace-Tang))
* Trick skip always failed tests [\#1195](https://github.com/alibaba/pouch/pull/1195) ([Ace-Tang](https://github.com/Ace-Tang))

### New Contributors

Here is the list of new contributors:

* [rhinoceros](https://github.com/rhinoceros)
* [soarpenguin](https://github.com/soarpenguin)
* [chuanchang](https://github.com/chuanchang)
* [raoqi](https://github.com/raoqi)
* [u2takey](https://github.com/u2takey)
* [shaloulcy](https://github.com/shaloulcy)
* [xiechengsheng](https://github.com/xiechengsheng)

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
* Separate integration test and cri test in travisCI [\#1098](https://github.com/alibaba/pouch/pull/1098) ([Letty5411](https://github.com/Letty5411))
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
* [dbdd4us](https://github.com/dbdd4us)
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

* With this PR, we can get the error informations when stream server handles `exec` or `attach` commands occurred errors [\#1007](https://github.com/alibaba/pouch/pull/1007)
* Add websocket support for cri stream server [\#985](https://github.com/alibaba/pouch/pull/985)
* Fix handle image format 'namespace/name:tag' correctly [\#981](https://github.com/alibaba/pouch/pull/981)
* Fix pull image and get its status with RefDigest [\#973](https://github.com/alibaba/pouch/pull/973)
* Store sandbox config informations for cri manager [\#955](https://github.com/alibaba/pouch/pull/955)
* Separate stdout & stderr of container io and support host network mode for sandbox [\#945](https://github.com/alibaba/pouch/pull/945)
* Implement ReadOnlyRootfs and add `no-new-privileges` support to cri manager [\#935](https://github.com/alibaba/pouch/pull/935)
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
