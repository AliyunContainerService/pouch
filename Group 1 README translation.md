[原文章地址](https://github.com/alibaba/pouch/blob/master/README.md)

## 主链接 

- <a href="#1">介绍</a>
- <a href="#2">特征</a> 
- <a href="#3">架构</a>
- <a href="#4">优势</a>
- <a href="#5">入门</a>
- <a href="#6">通过PouchContainer配置Kubernetes</a>
- 用户手册
  - [CLI手册](https://github.com/alibaba/pouch/tree/master/docs/commandline)
  - [API手册](https://github.com/alibaba/pouch/tree/master/docs/api)
- <a href="7">贡献</a>

## <a name="1">介绍</a>

PouchContainer 是由阿里巴巴集团创建的一款开源项目，旨在推动容器技术的发展。

PouchContainer 的愿景是推动容器生态系统，促进容器标准[OCI（Open Container Initiative）](https://github.com/opencontainers)，从而让容器技术成为云时代应用开发的基础。
PouchContainer 可以对任何应用程序进行打包，交付和运行，并通过强大隔离和最小开销为应用提供轻量级运行环境。PouchContainer 将不同运行环境中的应用隔离开来，并最大限度地减小操作工作量PouchContainer 最大限度地减小了应用开发人员编写云原生应用或者将旧版本应用程序迁移到云平台的工作量。

## <a name="2">特征</a>

PouchContainer最重要的特征有：

- **丰富的容器**：除了常用的运行容器的方法，PouchContainer包含了一个`rich container`的模式，它集成了更多的服务，钩子和其他很多容器内部组件，以保证容器的正常运行。
- **强大的隔离**：PouchContainer默认设置为安全，它包含非常多安全功能，比如基于管理程序的容器技术，lxcfs，目录磁盘配额，打补丁的Linux内核等等.
- **P2P分布**：PouchContainer利用基于P2P的分布式系统[Dragonfly](https://github.com/alibaba/dragonfly)，实现企业大规模级的快速容器图像分布。
- **内核兼容性**：使得兼容OCI的运行同样能够在像Linux内核2.6.32+等旧内核版本上运行。
- **标准兼容性**：PouchContainer不断拥抱容器生态系统来支持行业标准，比如CNI，CSI等等。
- **Kubernetes兼容性**：PouchContainer本身实现了Kubernetes容器运行时接口（CRI），从其他Kubernetes容器运行时迁移到PouchContainer会很顺利。

## <a name="3">架构</a>
我们从两个维度来描述PouchContainer的架构：说明PouchContainer如何适应于容器生态系统的**生态系统架构**和描述PouchContainer中不同组件间交互的**组件架构**。更多细节，请参见文件[architecture.md](https://github.com/alibaba/pouch/blob/master/docs/architecture.md)。

## <a name="4">优势</a>
相较于VM技术，PouchContainer有很多优势，其中最令人印象深刻的两个是**资源利用**和**应用中心**。
### 资源利用
PouchContainer显著提高了资源利用率：

- PouchContainer与OCI图像规范兼容。应用程序可以通过分层图像结构最小化存储使用。
- 增量图像分布节省了数据中心带宽消耗。
- 相较于基于VM的技术，运行时开销明显更小。
### 应用中心
PouchContainer为应用开发提供了一个更“以应用为中心”的方法：

- PouchContainer通过内核支持下和内核模式外的尖端技术，在应用程序之间提供强大的运行时隔离。
- PouchContainer支持跨平台和跨OS应用程序的交付。
- PouchContainer支持标准化应用程序图像规范，从而使得应用程序的分享和重复使用对开发人员和运营商来说变得非常容易。
## <a name="5">入门</a>
您可以轻松设置基本的PouchContainer环境，请参见 [INSTALLATION.md](https://github.com/alibaba/pouch/blob/master/INSTALLATION.md)。您需要在启动`pouchd`之前安装一些启动容器管理服务的软件包。您可以通过`pouch` CLI或RPC调用获取该服务。有关更多详细信息，请参见[CLI手册](https://github.com/alibaba/pouch/tree/master/docs/commandline)和[API手册](https://github.com/alibaba/pouch/tree/master/docs/api)。
## <a name="6">通过PouchContainer配置Kubernetes</a>
当您的机器安装了PouchContainer后，也许就来到了通过PouchContainer为Kubernetes集群赋能的激动时刻。您可以通过一个简单地指南快速体验这个惊人的组合[Kubernetes + PouchContainer](https://github.com/alibaba/pouch/blob/master/docs/kubernetes/pouch_with_kubernetes_deploying.md)。
## <a name="7">贡献</a>
欢迎您对PouchContainer进行修改，我们准备了一份详细的指南[CONTRIBUTING.md](https://github.com/alibaba/pouch/blob/master/CONTRIBUTING.md)。
## 问题
有关常见问题的详细信息（FAQ），请参见文件[FAQ.md](https://github.com/alibaba/pouch/blob/master/FAQ.md)。
## 路线图
有关路线图的更多详细信息，请参见文件[ROADMAP.md](https://github.com/alibaba/pouch/blob/master/ROADMAP.md)。
## 联系我们
我们鼓励您通过GitHub问题或提取请求来进行沟通。将来，如有必要，我们会提供更多沟通渠道。如果您有紧急问题，请通过[pouch-dev@list.alibaba-inc.com](mailto:pouch-dev@list.alibaba-inc.com)与PouchContainer团队联系。
## 许可证
PouchContainer通过Apache许可证2.0版获得许可。有关完整许可文本，请见[LICENSE](https://github.com/alibaba/pouch/blob/master/LICENSE)。
