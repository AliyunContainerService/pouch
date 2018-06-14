# Frequently Asked Questions

## What is PouchContainer project

PouchContainer is a tool providing container services for developers and operators. It helps to build a successful work flow for IT engineering teams. PouchContainer also accelerates DevOps innovations for enterprises.

For developers, PouchContainer provides a standard way to package their applications. With PouchContainer, developers can pack application with little effort, and run it everywhere. The standardized environment provided by PouchContainer makes it easier to run Continuous Integration, and improves the efficiency of Continuous Delivery.

For operators, PouchContainer enables automation, and largely reduces manual operations. With PouchContainer, operators no longer worry about the heterogeneous machine architecture and environments. PouchContainer enables them to focus more on application operation.

For data center owners, PouchContainer is the best choice you ever have. Compared to VM technology, PouchContainer helps increase the resource utilization with the similar level of isolation.

## Why is it named PouchContainer

PouchContainer refers to some kinds of small bags. One kind is brood pouch which is used to protect very young life. This is a metaphor that Software PouchContainer has its responsibility to take care of applications very closely. In another word, application is the keyword in PouchContainer's world.

## What is the history of PouchContainer

Originally in 2011, PouchContainer is a pure container service in Alibaba. It is used to serve millions of trade business of Taobao. At that time, PouchContainer is based on a technology named by [LXC](https://en.wikipedia.org/wiki/LXC).

With the evolution of container technology in industry, [Docker](https://www.docker.com/) technology comes up and becomes popular with its innovative layered image technology. In 2015, PouchContainer introduces docker's images technology to its own architecture to make itself much stronger.

As more and more scenes experience, PouchContainer gets lots of polishes and definitely turns production-ready. Currently it supports most of the running of business in Alibaba.

## What is the role of PouchContainer in container ecosystem

Maybe many people would say that container ecosystem has been very mature. What is the role of PouchContainer?

First, we admit there are so many software in container ecosystem. However, according to container technology experience in Alibaba, current ecosystem is good, but can be better, especially on the attitude towards application as container engine. So PouchContainer is a lighter and more useful container engine in ecosystem.

In the underlying support of container runtime, PouchContainer takes such opinion that lighter VM based on hypervisor is as important as container based on kernel support, such as cgroup and namespace. We can say container engine part of PouchContainer is very pure. More responsibility on container orchestration relies on upper orchestration technologies, like [Kubernetes](https://github.com/kubernetes/kubernetes), [Mesos](https://github.com/apache/mesos).

## What is difference between PouchContainer and Docker

PouchContainer and Docker are both excellent container solution for users. They do similar things if comparing them at a glance. But more specifically, they have different emphasize on each one's target. PouchContainer takes more emphasis on application experience, while Docker advocates "one process one container" a lot. PouchContainer cannot ignore isolation threat of container technology in some particular scenes, while Docker relies on kernel to achieve isolation heavily. PouchContainer brings an open attitude for the surrounding ecosystem, while docker also works on this but maybe not so much.

Here we list some additional features of PouchContainer:

* rich container: It means that there is not only one application process in container any more. Each container has its init process, and other system services on premise according to user's need.
* strong isolation: PouchContainer can create a VM with hypervisor technology via [runV](https://github.com/hyperhq/runv) and [clearcontainer](https://github.com/clearcontainers/runtime)
* high kernel compatibility: PouchContainer has a wide range of kernel version support. It is a long road for industry to upgrade kernel version to 3.10+. PouchContainer could help legacy kernel world to enjoy the fresh container technology.
* P2P image distribution: In a very large datacenter, image distribution is heavy load for network. PouchContainer can take advantage of P2P image distribution solutions to improve this.

## What is the difference between PouchContainer and Kubernetes

Kubernetes is an open source project for managing containerized applications across multiple hosts, providing basic mechanisms for deployment, maintenance, and scaling of applications. While PouchContainer mainly focuses on container management with a rich container runtime diversity. For clearer relationship between PouchContainer and Kubernetes, you can refer to [Ecosystem Architecture](docs/architecture.md#ecosystem-architecture).

If you are choosing overall container solution for a 100+ nodes cluster wide scenario, along with application maintenance friendly with less effort, Kubernetes is one which you can choose. If you wish to harden you container's running, pack your application efficiently and standardize your heterogeneous infrastructure with a unified operation experience, PouchContainer is what you want.

What is more, in data center's architecture, Kubernetes and PouchContainer locate in different layers. We can say that Kubernetes is in the upper layer towards PouchContainer. Actually we can combine Kubernetes and PouchContainer in a very efficient way. PouchContainer takes a role of runtime solution, and Kubernetes plays a role of orchestration. PouchContainer takes over fine-grained resource from infrastructure, like CPU, memory, network, disk and so on. Then it provides these resource metrics to upper Kubernetes for scheduling usage. When Kubernetes runs to meet application's maintenance demand, it can pass the request down to PouchContainer to provide secure and isolated container carriers for applications.

## What is PouchContainer's rich container

Rich container is a very useful container mode when containerizing applications. This mode helps technical staff to complete packaging fat applications almost with no effort. It provides efficient ways to equip more basic software or system services except for target application in a single container . Then applications in containers could be running as smoothly as usual in VM or physical machine. This is a more generalized application-centric mode. This mode brings no invasiveness at all to both developers and operators. Especially for operators, they could have abilities to maintain applications in container with all essential tools or service processes they may need as usual.

Rich container mode is not the default mode PouchContainer provides. It is an additional mode PouchContainer brings to extend users' container experience. Users can still manage ordinary containers by switching rich container flag off. For more info, please refer to [rich container](./docs/features/pouch_with_rich_container.md).

## What is relationship between containerd and PouchContainer

To be added.

## What is the image storage driver PouchContainer uses

PouchContainer uses overlay2 as its default image storage driver. To be more detailed, containerd 1.0.0+ begins to make use of overlay2 to store images. While currently containerd is a base component in PouchContainer. For Linux mainline kernel, only kernel version 4.0+ supports overlay2. So, in theory, PouchContainer can only run on Linux distribution which are based on kernel 4.0+. However, RedHat series, like CentOS 7.2+ and so on, supports overlay2 stably since kernel version 3.10.327. For more information on RedHat, please refer to [Red Hat Enterprise Linux Release Dates](https://access.redhat.com/articles/3078#RHEL7).

## Does PouchContainer support Kubernetes

Yes, PouchContainer has already implemented all the CRI(container runtime interface). You can use PouchContainer under Kubernetes out of box. For more details, please refer to [Kubernetes](./docs/kubernetes).

## Does PouchContainer support Mesos/DCOS

To be added.

## What is version rule of PouchContainer

We set the version rule of PouchContainer on the basis of [SemVer](http://semver.org/). Briefly speaking, we follow version number of MAJOR.MINOR.PATCH, such as 0.2.1, 1.1.3. For more details, please refer to [SemVer](http://semver.org/).

## What is the roadmap of PouchContainer

See [ROADMAP.md](./ROADMAP.md)

## How to contribute to PouchContainer

It is warmly welcomed if you have interest to contribute to PouchContainer.

More details, please refer to [CONTRIBUTION.md](./CONTRIBUTING.md)
