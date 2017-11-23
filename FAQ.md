# Frequently Asked Questions

## What does Pouch mean?

Pouch refers to some kinds of small bags. One kind is brood pouch which is used to protect very young life. This is a metaphor that Software Pouch has its responsibilty to take care of applications very closely. In another word, application is the keyword in Pouch's world.

## What can Pouch bring you?

Pouch is a convenient tool providing container services for developers and operators. 

Pouch can help to build a successful work flow for your IT engineering easily. We can also have such idea that Pouch is a huge helper to inovate DevOps for enterprises.

For developers, it can provide a standard way to package various applications. The standarized environment provided by Pouch could help you easily run CI(continuous integration) and improve CD(continuous delivery) efficieny. With this software, engineers can pack apllication with no effort and run applications out of box.

For operators, automation is one of the obvious benifits. Manual operation can be reduced to a fairly small percentage with Pouch. Operator could never mind the heterogeneous machine architeture and operation system. And they have alibities to focus more on application operation rather than hardware operation.

In addition, if you own a huge datacenter, Pouch is the best choice you ever have. It can increase the resource utilization of datacenter a lot at a very low effort. Besides, isolation ability is bright feature of Pouchd.

## What is the history of Pouch?

Originally in 2011, Pouch is a pure container service in Alibaba. It is used to serve millions of trade business of Taobao. At that time, Pouch is based on a technology named by [LXC](https://en.wikipedia.org/wiki/LXC). 

With the evolution of container technology in industry, [Docker](https://www.docker.com/) technology comes up and becomes popular with its inovative layered image technology. In 2015, Pouch introduces docker's images technology to its own architeture to make itself much stronger.

As more and more scenes experience, Pouch gets lots of polishes and denifitely turns production-ready. Currently it supports most of the running of business in Alibaba.

## What is the role of Pouch in container ecosystem?

Maybe many people would say that container ecosystem has been very mature. What is the role of Pouch?

First, we admit there are so many software in container ecosystem. However, according to container technology experience in Alibaba, current ecosystem is good, but can be better, especially on the attitude towards application as container engine. So Pouch is a lighter and more useful container engine in ecosystem. 

In the underlying support of container runtime, Pouch takes such opinion that lighter VM based on hypervisor is as important as container based on kernel support, such as cgroup and namespace. We can say container engine part of Pouch is very pure. More responsibility on container orchestration relies on upper orchestration technologies, like [Kubernetes](https://github.com/kubernetes/kubernetes), [Mesos](https://github.com/apache/mesos).

## What is difference between Pouch and Docker?

Pouch and Docker are both excellent container solution for users. They do similar things if comparing them at a glance. But more specifically, they have different emphasize on each one's target. Pouch takes more emphasis on application experience, while Docker advocates "one process one container" a lot. Pouch cannot ignore isolation threat of container technology in some particular scene, while Docker relies on kernel to achieve isolation heavily. Pouch brings an open attitude for the surrounding ecosystem, while docker also works on this but maybe not so much. 

Here we list some additional features of Pouch: 

* rich container: It means that there is not only one application process in container any more. Each container has its init process, and other systsem services on premise according to user's need.
* strong isolation: Pouch can create a VM with hypervisor technology via [runV](https://github.com/hyperhq/runv) and [clearcontainer](https://github.com/clearcontainers/runtime)
* high kernel compatibility: Pouch has a wide range of kernel version support. It is a long road for industry to upgrade kernel version to 3.10+. Pouch could help legecy kernel world to enjoy the fresh container technology.
* P2P image distribution: In a very large datacenter, image distribution is heavy load for network. Pouch can take advantage of P2P image distribution solutions to improve this.

## What is version rule of Pouch ?

We set the version rule of Pouch on the basis of [SemVer](http://semver.org/). Briefly speaking, we follow version number of MAJOR.MINOR.PATCH, such as 0.2.1, 1.1.3. For more details, please refer to [SemVer](http://semver.org/).

## What is the roadmap of Pouch?

See [ROADMAP.md](./ROADMAP.md)

## How to contribute to Pouch?

It is warmly welcomed if you have interest to contribute to Pouch.

More details, please refer to [CONTRIBUTION.md](./CONTRIBUTING.md)
