
# Pouch - An Efficient Enterprise-class Rich Container Engine

[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
[![GoDoc](https://godoc.org/github.com/alibaba/pouch?status.svg)](https://godoc.org/github.com/alibaba/pouch)
[![Build Status](https://travis-ci.org/alibaba/pouch.svg?branch=master)](https://travis-ci.org/alibaba/pouch)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bhttps%3A%2F%2Fgithub.com%2Falibaba%2Fpouch.svg?type=shield)](https://app.fossa.io/projects/git%2Bhttps%3A%2F%2Fgithub.com%2Falibaba%2Fpouch?ref=badge_shield)
[![Go Report Card](https://goreportcard.com/badge/github.com/alibaba/pouch)](https://goreportcard.com/report/github.com/alibaba/pouch)

![pouchcontainer-logo-800](https://user-images.githubusercontent.com/6755791/39180769-55f3b742-47ea-11e8-8762-78aeedcbba78.png)

## Main Links

- [Introduction](#introduction)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Deploying Kubernetes with Pouch](#deploying-kubernetes-with-pouch)
- User Manual
  - [CLI Manual](docs/commandline)
  - [API Manual](docs/api)
- [Contributing](#contributing)

## Introduction

Pouch is an open-source project created by Alibaba Group to promote the container technology movement.

Pouch's vision is to advance container ecosystem and promote container standards [OCI(Open Container Initiative)](https://github.com/opencontainers), so that container technologies become the foundation for application development in the Cloud era.

Pouch can pack, deliver and run any application. It provides applications with a lightweight runtime environment with strong isolation and minimal overhead. Pouch isolates applications from varying runtime environment, and minimizes operational workload. Pouch minimizes the effort for application developers to write Cloud-native applications, or to migrate legacy ones to a Cloud platform.

## Features

Pouch's most important features are:

- **Rich container**: Besides the common ways of running container, Pouch includes a `rich container` mode, which integrates more services, hooks, and many others container internals to guarantee container's running like usual.
- **Strong isolation**: Pouch is designed to be secure by default. It includes lots of security features, like hypervisor-based container technology, lxcfs, directory disk quota, patched Linux kernel and so on.
- **P2P distribution**: Pouch utilizes [Dragonfly](https://github.com/alibaba/dragonfly), a P2P-base distribution system, to achieve lightning-fast container image distribution at enterprise's large scale.
- **Kernel compatibility**: Enables OCI-compatible runtimes to work on old kernel versions, like linux kernel 2.6.32+.
- **Standard compatibility**: Pouch keeps embracing container ecosystem to support industry standard, such as CNI, CSI and so on.
- **Kubernetes compatibility**: Pouch has natively implemented Kubernetes Container Runtime Interface(CRI). It will be smooth to migrate from other Kubernetes container runtime to Pouch.

## Architecture

We describe Pouch's architecture from two dimensions: **ecosystem architecture**  which illustrates how Pouch fits into the container ecosystem and **component architecture** which describes the interactions between various components inside Pouch. For more details, please refer to file [architecture.md](docs/architecture.md).

## Advantages

Pouch has lots of advantages over VM technologies. Two of the most impressive ones are **Resource Utilization** and **Application Centric**.

### Resource Utilization

Pouch significantly improves resource utilization:

- Pouch is compatible with OCI image spec. Applications can minimize their storage usage with layered image structure.
- Incremental image distribution, saves datacenter bandwidth consumption.
- Significantly less runtime overhead than VM-based technologies.

### Application Centric

Pouch offers a more "application centric" approach for application development:

- Pouch provides strong runtime isolation between applications, with cutting-edge technology both within kernel support and beyond kernel mode.
- Pouch enables cross-platform and cross-OS application delivery.
- Pouch supports standardized application image spec, so application sharing and reusing becomes trivial for developers and operators.

## Getting Started

You can easily setup a basic Pouch environment, see [INSTALLATION.md](INSTALLATION.md). You'll need to install a few packages before starting `pouchd`, which starts a container management service. The service can be accessed through the `pouch` CLI or RPC calls. For more details, please refer to [CLI Manual](docs/commandline) and [API Manual](docs/api).

## Deploying Kubernetes With Pouch

After installing Pouch on your machine, maybe it is the exiciting moment for you to power your Kubernetes cluster by Pouch. There is an easy guide for you to quickly experience this amazing combination [Kubernetes + Pouch](docs/kubernetes/pouch_with_kubernetes_deploying.md).

## Contributing

You are warmly welcomed to hack on Pouch. We have prepared a detailed guide [CONTRIBUTING.md](CONTRIBUTING.md).

## FAQ

For more details about frequently asked questions (FAQ), please refer to file [FAQ.md](FAQ.md).

## Roadmap

For more details about roadmap, please refer to file [ROADMAP.md](ROADMAP.md).

## Connect with us

You are encouraged to communicate everything via GitHub issues or pull requests. In the future, we would provide more channels for communication if necessary.

If you have urgent issues, please contact Pouch team at [pouch-dev@list.alibaba-inc.com](mailto:pouch-dev@list.alibaba-inc.com).

## License

Pouch is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.
