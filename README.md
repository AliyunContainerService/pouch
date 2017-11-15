
## ![Pouch](docs/logo/pouch_10x4_orange.png)

## Links

- [Introduction](#introduction)
- [Advantanges and Disadvantages](#advantages)
- [Docs](docs)
    - [Installation](#installation)
    - [Usage Guide](docs/commandline)
    - [Architecture Design](docs/architecture.md)
    - [Contributing](CONTRIBUTING.md)
- [FAQ](FAQ.md)
- [ROADMAP](ROADMAP.md)
- [LICENSE](LICENSE)

## Introduction

Pouch is an open-source project created by Alibaba Group to promote the container technology movement. 

Pouch can pack, deliver and run any application. It provides the enviroment for applications with strong isolation in quite lightweight way. Pouch not only splits the application itself from the underlying environment, but also has ability to remain the good experience of operation.

The ambitious of Pouch is embracing container ecosystem and perfecting container standards [OCI(Open Container Initiative)](https://github.com/opencontainers). With the ability Pouch provides, people can spend less energy transforming applications to be cloud native. 

## Advantages

Pouch has lots of advantages over VM technologies. Two of the most impressive ones are **Resource Utilization** and **Application Centric**.

### Resource Utilization

Pouch improves resource utilization of application with significant effect:

* Image technology of Pouch is compatile with OCI image spec. It could help application take minimal space of storage with layerd image originazation.
* Via incremental way provided by image when application distribution, datacentre bandwidth reource could be saved a lot.
* Unlike original VM technology, the auxiliary resource which is needed for applications' boot could be minimal, while for VM there are resources spared for kernel and system services and so on.

### Application Centric

Pouch pays more emphasis on view of application, and we can call this "application centric":

* Application needs an absolutely isolated environment. Pouch provides strong isolation for them with cutting-edge technology both within kernel support and beyond kernel mode.
* Application delivery turns to be out of box. Pouch improve the portability of application, no matter cross-platform or cross-os.
* Application delivery period should be minimal. Pouch shorted this by standardizing the application image spec between developers and operators.

## Installation

See [INSTALLATION.md](INSTALLATION.md).

## FAQ
File [FAQ.md](FAQ.md) contains frequently asked question (FAQ).

## Roadmap

See [ROADMAP.md](ROADMAP.md).

## License

Pouch is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.
