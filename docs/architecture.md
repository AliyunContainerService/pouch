# Architecture

To clarify standpoint of PouchContainer in container system, we construct it with explicit architecture. To ensure the clear separation of functionality, we have organized PouchContainer with components. Therefore, when mentioning architecture, we usually include two parts:

* ecosystem architecture
* component architecture

## Ecosystem Architecture

In PouchContainer's roadmap, we set ecosystem embracing as a big target. To upper orchestrating layer, PouchContainer supports Kubernetes and Swarm. To underlying runtime layer, PouchContainer is compatible with oci-compatible runtime, such as [runC](https://github.com/opencontainers/runc), [runV](https://github.com/hyperhq/runv), runlxc and so on. To make storage and network big supplements, [CNI](https://github.com/containernetworking/cni) and [CSI](https://github.com/container-storage-interface) are in scope right there.

![Ecosystem Architecture](static_files/pouch_ecosystem_architecture_no_logo.png)

The ecosystem architecture may be a little bit complicated at first glance. Take it easy. We can get a thorough understanding of it from the following three dimensions.

### Runtime Layer

Runtime layer is located on right-top in the architecture picture. This dimension mainly focus on OCI-compatible runtimes supported in PouchContainer. These runtimes unify specifications for standards on operating system process and application containers. Currently, PouchContainer supports four kinds of OCI-compatible runtimes:

* runC
* runlxc
* runV
* clear containers

With runC, PouchContainer creates common containers like other container engine does, for example docker. With runlxc, PouchContainer creates containers based on LXC. runlxc helps a lot when users need to run containers on a wide variety of Linux kernels with the ability to be compatible with kernel 2.6.32+. Hypervisor-based containers have many application scenarios as well. PouchContainer will support it with runV and clear container.

All these four runtimes mentioned above are supported under containerd. Containerd takes over all detailed container management, including creation, start, stop, deletion and so on.

### Orchestration Layer

PouchContainer is always active on supporting Kubernetes since the first day when it is designed. We illustrate this part on the top half of the architecture picture. First, PouchContainer will integrate cri-containerd inside, so Kubernetes can easily dominate PouchContainer to manage Pod. The workflow will pass cri-containerd, containerd client, containerd, runC/runV and pod. When configuring network of Pod, cri-containerd will take advantage of network plugins which implement CNI interface.

### Container Layer

We support not only Pod in Kubernetes cluster, but also simple container management for users. This is especially useful for developers. In another word, PouchContainer supports single container API. In this way, workflow passes pouchd, containerd client, containerd, runC/runV and container. On the aspect of network, PouchContainer uses libnetwork to construct container's network. What's more, lxcfs is also used to guarantee the isolation between containers and between containers and host.

## Component Architecture

Ecosystem architecture of PouchContainer shows the location of itself in the container ecosystem. The following picture shows the component architecture of PouchContainer. In component architecture, we divide PouchContainer into two main parts: PouchContainer CLI and Pouchd.

![Component Architecture](static_files/pouch_component_architecture.png)

### PouchContainer CLI

There are lots of different commands encapsulated in PouchContainer CLI, like create, start, stop, exec and so on. Users can interact with Pouchd by PouchContainer CLI. When executing a command, PouchContainer CLI will translate it into Pouchd API calls to satisfy users' demand. PouchContainer Client API is a well-encapsulated package in PouchContainer CLI. It is very easy for others to integrate PouchContainer Client Package into third-party software. And this package currently only supports Golang language. When calling Pouchd via PouchContainer Client Package, the communication is over HTTP.

### Pouchd

Pouchd is designed decoupled from the very beginning. It makes Pouchd quite easy to understand. And it helps a lot for us to hack on PouchContainer. In general, we treat that Pouchd can be split into the following pieces:

* HTTP server
* bridge layer
* Manager(System/Network/Volume/Container/Image)
* ctrd

**HTTP Server** receives API calls directly and replies to client side. Its job is to parse requests and construct correct struct which is supposed to be passed to bridge layer, and to construct response no matter server succeeds in handling request or fails.

**bridge layer** is a translation layer which handles objects from client to meet managers or containerd's demand and handles objects from managers and containerd to make response compatible with Moby's API.

**Manager** is main processor of Pouchd. It deals proper object from requests, and does the corresponding work. There are five managers currently in Pouchd: container manager, image manager, network manager, volume manager and system manager.

**ctrd** is containerd client in Pouchd. When managers need to communicate with containerd, ctrd is the right thing to do this work. Managers call functions in ctrd and send request towards containerd. In addition, when state of container changes, containerd is the first component to be aware of this, and ctrd has container watch goroutines to detect this and update inner data stored in cache.
