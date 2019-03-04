# CRI API CHANGELOG

* [Overview](#overview "Overview")
* [Requirements](#requirements "Requirements")
* [The Changes Of CRI API](#the-changes-of-cri-api "The changes of CRI API")
    * [UpdateContainerResources](#updatecontainerresources "UpdateContainerResources()")
    * [ContainerStatus](#containerstatus "ContainerStatus()")
    * [ImageStatus](#imagestatus "ImageStatus()")
    * [CreateContainer](#createcontainer "CreateContainer()")
    * [RemoveVolume](#removevolume "RemoveVolume()")
    * [StartPodSandbox](#startpodsandbox "StartPodSandbox()")
* [Pull Request](#pull-request)

## Overview

Because the CRI interface of Kubernetes cannot meet the customized development of Kubelet at present, it is necessary to provide the required functions of Kubelet by extending the API of CRI. On top of CRI's existing functionality, use the field extension or add method to make CRI meet the requirements.

Kubernetes Version: V1.10.0+

## Requirements

1. Support the ability to update the container diskquota.
    - Scenario:
        - Limits the directory size within the container.
2. Support the ability to acquire the volumes of the Image.
    - Scenario:
        - In the Upgrade process, the volumes and mountpoints of the old container need to be read. At the same time, the volumes in the new image also need to be read. If the mount point is the same, the volumes in the new image cover the volumes in the old container.
3. Support to the ability to acquire the volumes of the Container.
    - Scenario:
        - In the Upgrade process, the volumes of the new container needs to remain consistent with that of the old container, so the volumes of the old container needs to be read.
4. Support to the ability to acquire the envs of the Container.
     - Scenario:
        - In the Upgrade process, the env of the new container needs to remain consistent with that of the old container, so the envs of the old container needs to be read.
5. Support to the ability to start the PodSandbox specified and setup the network.
     - Scenario:
        - It will fail to get IP after PodSandbox restarts  because of the external factors such as shutdown the host.
  
# The Changes Of CRI API

## UpdateContainerResources

### What To Solve?

1. Support the ability to update the container quotadir.
2. Support more container resource fields define in OCI spec.
3. Support the ability to update container runtime spec annotations.

### Modification

+ Add the `DiskQuota` field in `LinuxContainerResources`, referring to the definition of Resources in [moby](https://github.com/moby/moby/blob/master/api/types/container/host_config.go) for better compatibility. The new fields are as follows：
+ Add some fields from [OCI spec](https://github.com/opencontainers/runtime-spec/blob/master/specs-go/config.go)

The changes need to be made in the proto file are as follows:

```
message UpdateContainerResourcesRequest {
    // ID of the container to update.
    string container_id = 1;
    // Resource configuration specific to Linux containers.
    LinuxContainerResources linux = 2;
    // Annotations contains arbitrary metadata for the container.
    // Note: spec_annotations is the metadata for container runtime, such as runc. Not
    // the 'annotations' for Kubernetes objects.
    map<string, string> spec_annotations = 3;
}

message LinuxContainerResources {
    ......
    //***New Fields***//
    // DiskQuota constrains the disk
    map<string,string> disk_quota = 100;
    // Block IO weight (relative weight vs. other containers)
    uint32  blkio_weight = 101;
    repeated WeightDevice blkio_weight_device = 102;
    repeated ThrottleDevice blkio_device_read_bps = 103;
    repeated ThrottleDevice blkio_device_write_bps = 104;
    repeated ThrottleDevice blkio_device_read_IOps = 105;
    repeated ThrottleDevice blkio_device_write_IOps = 106;
    // Kernel memory limit (in bytes)
    int64 kernel_memory = 107;
    // Memory soft limit (in bytes)
    int64 memory_reservation = 108;
    // Tuning container memory swappiness behaviour
    Int64Value memory_swappiness = 109;
    // List of ulimits to be set in the container
    repeated Ulimit ulimits = 110;
}

// WeightDevice is a structure that holds device:weight pair
message WeightDevice {
    // Path of weightdevice.
    string path = 1;
    // Weight of weightdevice.
    uint32 Weight = 2;
}

// ThrottleDevice is a structure that holds device:rate_per_second pair
message ThrottleDevice  {
    // Path of throttledevice.
    string path = 1;
    // Rate of throttledevice.
    uint64 rate = 2;
}

//  Ulimit is a human friendly version of Rlimit.
message Ulimit {
    // Name of ulimit.
    string name = 1;
    // Hard limit of ulimit.
    int64 hard = 2;
    // Soft limit of Ulimit.
    int64 soft = 3;
}
```

## ContainerStatus

### What To Solve?

+ Support to the ability to acquire the volumes of the Container.
+ Support to the ability to acquire the Resource of the Container.
+ Pass the quotaID generated when the container is created for disk reuse.
+ Support to the ability to acquire the envs of the Container.

### Modification

+ The `ContainerStatus` struct is used only in `ContainerStatusResponse` in CRI, so the volumes of the container can be obtained by directly adding `volume` field  to the `ContainerStatus` struct.
+ Add Resources field to support the retrieval of container's resource.
+ When get ContainerStatus, the return object of `ContainerStatusResponse` will contain the field of `QuotaId` .
+ When get ContainerStatus, the return object of `ContainerStatusResponse` will contain the field of `Envs` .

```
// ContainerStatus represents the status of a container.
type ContainerStatus struct {
    ......
    //***New Fields***//
    // Volumes of container
    Volumes map[string]struct{} `protobuf:"bytes,100,opt,name=volumes,json=volumes" json:"volumes,omitempty"`
    // Resources specification for the container
    Resources *LinuxContainerResources `protobuf:"bytes,101,opt,name=resources" json:"resources,omitempty"` 
    // QuotaId of the container
    QuotaId string `protobuf:"bytes,102,opt,name=quota_id,json=quotaId,proto3" json:"quota_id,omitempty"`
    // List of environment variable to set in the container.
    Envs []*KeyValue `protobuf:"bytes,103,rep,name=envs" json:"envs,omitempty"`
}
```

The changes need to be made in the proto file are as follows:

```
// ContainerStatus represents the status of a container.
message ContainerStatus {
    ......
    //***New Fields***//
    // Volumes of container
    map<string,  Volume> volumes= 100;
    // Resources specification for the container
    LinuxContainerResources resources = 101;
    // QuotaId of the container
    string quota_id = 102;
    // List of environment variable to set in the container.
    repeated KeyValue envs = 103;
}

message Volume {
}
```

## ImageStatus

### What To Solve?

+ Support the ability to acquire the volumes of the Image.

### Modification

+ Add `volumes` field in the Image struct.

```
// Basic information about a container image.
type Image struct {
    ......
    //***New Fields***//
    // Volumes of image
    Volumes map[string]struct{} `protobuf:"bytes,7,opt,name=volumes,json=volumes" json:"volumes,omitempty"`
}
```

The changes need to be made in the proto file are as follows:

```
// Basic information about a container image.
message Image {
    ......
    //***New Fields***//
    // Volumes of image
    map<string, Volume> volumes= 100;
}
```

## CreateContainer

### What To Solve?

+ Support the ability to set DiskQuota.
+ Add missing fields.

### Modification

+ `LinuxContainerConfig` contains `LinuxContainerResources`, which have changed in UpdateContainerResources().So after changing LinuxContainerResources, the Create process already supports the setting of DiskQuota.
+ For missing fields are as follows (not all):
    + NetPriority : Set network priorities
+ QuotaId : When creating container, pass parameters of the DiskQuota and QuotaId. (When QuotaId is -1, QuotaId will be automatically generated)

```
type ContainerConfig struct {
    ......
    //***New Fields***//
    // NetPriority of the container
    NetPriority int64 `protobuf:"bytes,100,opt,name=net_priority" json:"net_priority,omitempty"` 
    // QuotaId of the container
    QuotaId string `protobuf:"bytes,101,opt,name=quota_id,json=quotaId,proto3" json:"quota_id,omitempty"`
}
```

The changes need to be made in the proto file are as follows:

```
message ContainerConfig {
    ......
    //***New Fields***//
    // NetPriority of the containeri
    int64  net_priority = 100;
    // QuotaId of the container
    string quota_id = 101;
}
```

## RemoveVolume

### What To Solve?

+ After kubelet performs upgrade, the container upgraded cannot delete the anonymous volume inherited.

### Modification

+ Provides an interface for removing volume.
+ The containerstatus interface supports querying volume by name.

The changes need to be made in the proto file are as follows:

```
service VolumeService {
 ​   // RemoveVolume volume an volume 
    rpc RemoveVolume(RemoveVolumeRequest) returns (RemoveVolumeResponse) {}
}​

message RemoveVolumeRequest {
    // name of the volume to remove
    string volume_name = 1;
}
message RemoveVolumeResponse {}
```

Add name field in the Mount struct：

```
// Mount specifies a host volume to mount into a container.
message Mount {
    ......
    //***New Fields***//
    // Name of volume
    string name = 100;  
}

```

## StartPodSandbox

### What To Solve?

+ StartPodSandbox restarts a sandbox pod which was stopped by accident and setup the network with network plugin.

### Modification

+ Provides an interface for starting PodSandbox specified.

The changes need to be made in the proto file are as follows:

```
service RuntimeService {
...
    // Start a sandbox pod which was forced to stop by external factors.
    // Network plugin returns same IPs when input same pod names and namespaces
    rpc StartPodSandbox(StartPodSandboxRequest) returns (StartPodSandboxResponse) {}
...
}

message StartPodSandboxRequest {
    // ID of the PodSandbox to start.
    string pod_sandbox_id = 1;
}

message StartPodSandboxResponse {}
```

## PauseContainer / UnpauseContainer

### What to Solve?

+ PauseContainer pause a container.
+ UnpauseContainer unpause a container.
+ Scenario: Under serverless situation, we may pre-allocate a batch of container which were ready to serve, waiting online. Using pause container to balance the resource cost and application start-up time.

### Modification

+ Extend the RuntimeService interface

The changes need to be made in the proto file are as follows:


```
service RuntimeService {
...
    // PauseContainer pauses the container.
    rpc PauseContainer(PauseContainerRequest) returns (PauseContainerResponse) {}
    // UnpauseContainer unpauses the container.
    rpc UnpauseContainer(UnpauseContainerRequest) returns (UnpauseContainerResponse) {}
...
}

message PauseContainerRequest {
    // ID of the container to pause.
    string container_id = 1;
}

message PauseContainerResponse {}

message UnpauseContainerRequest {
    // ID of the container to unpause.
    string container_id = 1;
}

message UnpauseContainerResponse {}
```

## Pull Request

+ feature: extend cri apis for special needs [#1617](https://github.com/alibaba/pouch/pull/1617)
+ feature: extend cri apis for remove volume [#2124](https://github.com/alibaba/pouch/pull/2124)
+ feature: extend cri apis for support quotaID [#2138](https://github.com/alibaba/pouch/pull/2138)
+ feature: extend cri apis for get envs [#2163](https://github.com/alibaba/pouch/pull/2163)
+ feature: extend cri apis for support StartPodSandbox [#2242](https://github.com/alibaba/pouch/pull/2242)
+ feature: extend cri apis for support pause/unpause container [#2623](https://github.com/alibaba/pouch/pull/2623)
