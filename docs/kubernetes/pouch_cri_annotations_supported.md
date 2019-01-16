# CRI ANNOTATIONS SUPPORTED CHANGELOG

* [Overview](#overview "Overview")
* [The Annotations Supported](#the-annotations-supported "The Annotations Supported")
  * [Runtime choosing](#runtime-choosing "Runtime choosing")
  * [LXCFS switcher](#lxcfs-switcher "LXCFS switcher")
  * [VM passthrough config](#vm-passthrough-config "VM passthrough config")
* [The container labels rule](#the-container-labels-rule "The container labels rule")
  * [Used by PouchContainer implementation](#used-by-pouchcontainer-implementation "Used by PouchContainer implementation")
  * [Generated from kubernetes spec](#generated-from-kubernetes-spec "Generated from kubernetes spec")

## Overview

Currently, PouchContainer has lots of advantages over other container runtimes, such as:

* resource view isolation via LXCFS
* runtime choosing for runc-based container or runv-based containers
* and so on

While actually in kubernetes, there is no related API to support these feature.

While for these extended features in Kubernetes, Kubernetes has a hiden way to support this: make user-defined parameters in annotations field in pod's definition.
When CRI manager deals the annotation details, it could pass these parameters to container manager, and container manager definitely implement these features very well.

So, we need to accomplish the following things:

* define the specific naming in annotations for each feature;
* implement the transformation in CRI manager and pass them to container manager.

## The Annotations Supported

| Requirement                       | Field definition                             | Supported Kubernetes Version | Pull Request                               |
|-----------------------------------|----------------------------------------------|------------------------------|-------------------------------------------|
| Runtime choosing | io.kubernetes.runtime  | V1.6 +  | https://github.com/alibaba/pouch/pull/1593 |
| LXCFS switcher | io.kubernetes.lxcfs.enabled | V1.10 +  | https://github.com/alibaba/pouch/pull/2210 |
| VM passthrough config swither| io.alibaba.pouch.vm.passthru | V1.10+ | https://github.com/alibaba/pouch/pull/2437 |
| VM passthrough IP | io.alibaba.pouch.vm.passthru.ip | V1.10+ | https://github.com/alibaba/pouch/pull/2437 |

NOTES: **Specify runtimes using `io.kubernetes.runtime` annotation is Deprecated**. It is recommended to use [RuntimeClass](https://kubernetes.io/docs/concepts/containers/runtime-class) which is a stable feature for selecting the container runtime configuration to use to run a pod’s containers.

### Runtime choosing

#### What To Solve

1. Support choosing the runtime for a container by making runtime choosing supported.

#### How to verify it

1. Prerequisites Installation, the runtime binaries you will use is necessary.

2. You should start pouchd with the configuration like this:

```
pouchd --enable-cri --cri-version v1alpha2 --add-runtime runv=runv >pouchd.log 2>&1  &
```

3. After setting up your kubernetes cluster, you can create a deployment like this :

```
# cat pouch-runtime.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pouch-runtime
  labels:
    pouch: pouch-runtime
spec:
  selector:
    matchLabels:
      pouch: pouch-runtime
  template:
    metadata:
      labels:
        pouch: pouch-runtime
      annotations:
        io.kubernetes.runtime: runv
    spec:
      containers:
      - name: pouch-runtime
        image: docker.io/library/nginx:latest
        ports:
        - containerPort: 80

# kubectl create -f pouch-runtime.yaml
```

4. Use command `pouch ps` to observe the runtime of the container and use command `uname -a` to view the system info.

```
# pouch ps
Name                                                                                                     ID       Status          Created          Image                                                                 Runtime
k8s_pouch-runtime_pouch-runtime-76c8d4d79b-6l5w7_default_47e8f918-b7c8-11e8-b238-42010a8c0004_0          59ce1d   Up 8 seconds    9 seconds ago    docker.io/library/nginx:latest                                        runv
k8s_POD_pouch-runtime-76c8d4d79b-6l5w7_default_47e8f918-b7c8-11e8-b238-42010a8c0004_0                    9492db   Up 21 seconds   29 seconds ago   registry.cn-hangzhou.aliyuncs.com/google-containers/pause-amd64:3.0   runv

// the system info for host
# uname -a
Linux k8s 4.15.0-1019-gcp #20~16.04.1-Ubuntu SMP Thu Aug 30 11:52:19 UTC 2018 x86_64 x86_64 x86_64 GNU/Linux

// the system info for the pod with runv
# kubectl exec -it pouch-runtime-76c8d4d79b-6l5w7 -- bash
$ uname -a
Linux pouch-runtime-76c8d4d79b-6l5w7 4.12.4-hyper #18 SMP Mon Sep 4 15:10:13 CST 2017 x86_64 GNU/Linux

```

### LXCFS switcher

#### What To Solve

1. Support resource view isolation via LXCFS in CRI Manager by making LXCFS configurable supported.

#### How to verify it

1. Prerequisites Installation and make sure your LXCFS service is running.

2. Enable pouchd LXCFS (with --enable-lxcfs flag).

3. After setting up your kubernetes cluster, you can create a deployment like this :

```
# cat pouch-lxcfs.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pouch-lxcfs
  labels:
    pouch: pouch-lxcfs
spec:
  selector:
    matchLabels:
      pouch: pouch-lxcfs
  template:
    metadata:
      labels:
        pouch: pouch-lxcfs
      annotations:
        io.kubernetes.lxcfs.enabled: "true"
    spec:
      containers:
      - name: pouch-lxcfs
        image: docker.io/library/busybox:latest
        command:
          - top
        resources:
          requests:
            memory: "256Mi"
          limits:
            memory: "256Mi"

# kubectl create -f pouch-lxcfs.yaml
```

4. View the results.

```
# pouch ps
Name                                                                                                           ID       Status       Created       Image                                                                 Runtime
k8s_pouch_pouch-5ddd8fc467-rmtcw_default_bc4b7972-b181-11e8-adae-42010a8c0003_0                                5391a9   Up 8 hours   8 hours ago   docker.io/library/busybox:latest                                      runc
k8s_POD_pouch-5ddd8fc467-rmtcw_default_bc4b7972-b181-11e8-adae-42010a8c0003_0                                  60a833   Up 8 hours   8 hours ago   registry.cn-hangzhou.aliyuncs.com/google-containers/pause-amd64:3.0   runc

# pouch exec k8s_pouch_pouch-5ddd8fc467-rmtcw_default_bc4b7972-b181-11e8-adae-42010a8c0003_0 cat /proc/meminfo
MemTotal:         262144 kB
MemFree:          261368 kB
MemAvailable:     261368 kB
......
```

### VM passthrough config

#### What To Solve

pass through some config to qemu

1. `io.alibaba.pouch.vm.passthru` specify whether a config should be passed through to qemu
2. `io.alibaba.pouch.vm.passthru.ip` specify the IP of the container.

## The container labels rule

### Used by PouchContainer implementation

| LabelKey | Usage |
| --- | --- |
| io.kubernetes.pouch.type | identify whether a container is a sandbox or a regular container |
| io.kubernetes.sandbox.id | attaching to a regular container specify which sandbox it belongs to |
| io.kubernetes.container.logpath | log path of the container |

### Generated from kubernetes spec

PouchContainer would generate some labels according to the annotations in spec by attaching the prefix `annotation.` to each annotation key.

For example, annotation `"io.kubernetes.container.restartCount": "0"` would be converted to label `"annotation.io.kubernetes.container.restartCount": "0"`.
