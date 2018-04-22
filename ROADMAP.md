# Roadmap

Roadmap provides detailed description of items that project pouch would decide to prioritize. This helps pouch contributors understand more about the evolving direction and whether a potential contribution is off the direction.

If the feature is not listed below, it does not mean that we will never take that into consideration. We always say that all contributions are warmly welcomed. Please understand that such kind of contributions may take a little bit more time for committers to review.

We designed three parts in roadmap:

* Container Regular Management
* Strong Isolation
* Open to Ecosystem

## Container Regular Management

We will polish user's experience on container management as the first important step. [Moby](https://github.com/moby/moby) has popularized container API standard in industry. And pouch will follow this API standard to provide container service. In addition, pouch will take more care of more aspects on how to run container on top of various isolation unit. Better experience on taking care of applications is in the scope as well.

## Strong Isolation

A lot of work has been done to improve container's security in industry. But container technology has not reached the target yet. Pouch will take more effect on strong isolation, no matter on software side or hardware side. Since security is the largest obstacle for technology to apply in production environment, pouch will improve isolation ability in the following areas: userspace lxcfs to isolate resource view, hypervisor based container, kvm-based container and so on.

## Enhancement to Ecosystem

For being open to container ecosystem, Pouch will be designed to be scalable. As a container engine, pouch will support pod and be able to integrate upper orchestraion layer with [kubernetes](https://github.com/kubernetes/kubernetes). For fundamental infrastructure management, pouch will embrace [CNI](https://github.com/containernetworking/cni) and [CSI](https://github.com/container-storage-interface). In the aspect of monitoring, logging and so on, Pouch takes an open role to approach cloud native.

