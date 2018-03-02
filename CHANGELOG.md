# CHANGELOG

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