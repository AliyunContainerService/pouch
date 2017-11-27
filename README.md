
Pouch - An Efficient Container Engine
================

[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
[![GoDoc](https://godoc.org/github.com/alibaba/pouch?status.svg)](https://godoc.org/github.com/alibaba/pouch)
[![Build Status](https://travis-ci.org/alibaba/pouch.svg?branch=master)](https://travis-ci.org/alibaba/pouch)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bhttps%3A%2F%2Fgithub.com%2Falibaba%2Fpouch.svg?type=shield)](https://app.fossa.io/projects/git%2Bhttps%3A%2F%2Fgithub.com%2Falibaba%2Fpouch?ref=badge_shield)
[![Go Report Card](https://goreportcard.com/badge/github.com/alibaba/pouch)](https://goreportcard.com/report/github.com/alibaba/pouch)

<img src="docs/static_files/logo/pouch_10x4_orange.png" width="480">

## Main Links

- [Introduction](#introduction)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- User Manual
    - [CLI Manual](docs/commandline)
    - [API Manual](docs/api)   
- [Contributing](#contributing)

## Introduction

Pouch is an open-source project created by Alibaba Group to promote the container technology movement. 

Pouch can pack, deliver and run any application. It provides the enviroment for applications with strong isolation in quite lightweight way. Pouch not only splits the application itself from the underlying environment, but also has ability to remain the good experience of operation.

The ambitious of Pouch is embracing container ecosystem and perfecting container standards [OCI(Open Container Initiative)](https://github.com/opencontainers). With the ability Pouch provides, people can spend less energy transforming applications to be cloud native. 

## Features

Pouch project has many important features listed below:

- **Security**: Designed to be secure by default. Include lots of security features, like hypervisor-based container technology, lxcfs, patched Linux kernel and so on.
- **P2P distribution**: Easy to integrate with [Dragonfly](https://github.com/alibaba/dragonfly) which is a P2P-base container image distribution system to significantly improve image distribution efficiency.
- **Rich container**: Besides the common way of container usage, also includes `rich container` mode which integrates more services, prestart hooks and more things inside container to guarantee container's running like usual.
- **Kernel compatibility**: Supports more OCI-compatible runtimes to support low version kernels, like linux kernel 2.6.32+.
- **Simple and easy to use**: Very few steps needed to setup Pouch. 

## Architecture

We describe Pouch's architecture from two dimensions: **ecosystem architecture**  which illustrates the place of Pouch in container ecosystem and **component architecture** which shows how we design decoupled components inside Pouch. For more details, please refer to file [architecture.md](docs/architecture.md).

## Advantages

Pouch has lots of advantages over VM technologies. Two of the most impressive ones are **Resource Utilization** and **Application Centric**.

### Resource Utilization

Pouch improves resource utilization of application with significant effect:

* Image technology of Pouch is compatible with OCI image spec. It could help application take minimal space of storage with layered image originization.
* Via incremental way provided by image when application distribution, datacentre bandwidth reource could be saved a lot.
* Unlike original VM technology, the auxiliary resource which is needed for applications' boot could be minimal, while for VM there are resources spared for kernel, system services and so on.

### Application Centric

Pouch pays more emphasis on view of application, and we can call this "application centric":

* Application needs an absolutely isolated environment. Pouch provides strong isolation for them with cutting-edge technology both within kernel support and beyond kernel mode.
* Application delivery turns to be out of box. Pouch improves the portability of application, no matter cross-platform or cross-os.
* Application delivery period should be minimal. Pouch shorted this by standardizing the application image spec between developers and operators.

## Getting Started

It is quite easy for us to start a journey to Pouch. Just install a few prerequisites before starting, and start pouchd with default configuration. Then you can enjoy container management both by API and pouch cli. For more details, please refer to file [INSTALLATION.md](INSTALLATION.md).

## User Manual

We can experience or hack Pouch in two ways: API side and CLI side. No matter which way you like, we provide detailed manuals for you. For more details, please refer to [CLI Manual](docs/commandline) and [API Manual](docs/api).

## Contributing

It is warmly welcomed if you are interested in hacking on Pouch. Here we have prepared detailed guidance [CONTRIBUTING.md](CONTRIBUTING.md) for you to follow. If you are not saitisfied with the guide, I think you know what to do. Yeah, contribute to improve it.

## FAQ

For more details about frequently asked questions (FAQ), please refer to file [FAQ.md](FAQ.md). 

## Roadmap

For more details about roadmap, please refer to file [ROADMAP.md](ROADMAP.md).

## Connect with us

At First, we have to say that it is encouraged to communicate everything via GitHub issues or GitHub pull requests, and all these should be in English. There is no difference between sending pouch members a regular email and pinging them on GitHub. In the future, we will provide more channels for community if necessary.

If you have ugent issues, please contact with this email address [pouch-dev@list.alibaba-inc.com](mailto:pouch-dev@list.alibaba-inc.com) which is mentioned in [CONTRIBUTING.md](CONTRIBUTING.md)

## License

Pouch is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.
