[原文章地址](https://github.com/alibaba/pouch/blob/master/ROADMAP.md)

# 路线图
路线图提供了PouchContainer项目将决定优先考虑的内容的详细说明。这有助于PouchContainer的贡献者更多地了解演化趋势以及潜在的贡献是否偏离方向。

如果有功能未在下面列出，并不意味着我们永远不会考虑这一点。我们一直说，所有的贡献都会受到热烈欢迎。请大家理解，此类贡献可能需要更多时间让提交者进行审核。

我们在路线图中设计了三个部分：

- 容器常规管理
- 强隔离
- 面向生态系统开放

## 容器常规管理

我们把优化用户在容器管理方面的体验作为第一个重要步骤。[Moby](https://github.com/moby/moby)已在工业中推广了容器API的标准。PouchContainer将遵循该API标准来提供容器服务。此外，PouchContainer将更多地关注如何在多元隔离单元上运行容器的更多方面问题。同时，处理应用程序的更好体验也在关注范围中。

## 强隔离

虽然在提高容器在工业中的安全性方面已经做了很多工作，但容器技术尚未达到目标。无论是在软件方面还是在硬件方面，PouchContainer都会对强隔离产生更大的影响。由于安全性是技术应用于生产环境的最大障碍，PouchContainer将提高以下领域的隔离能力：用于隔离资源视图的用户空间lxcfs，基于管理程序的容器，基于kvm的容器等。

## 加强生态系统
为了对容器生态系统开放，PouchContainer将被设计为可扩展。作为容器引擎，PouchContainer将支持集群并能够将上层编排层与[kubernetes](https://github.com/kubernetes/kubernetes)进行集成。对于基础设施管理，PouchContainer将采用[CNI](https://github.com/containernetworking/cni)和[CSI](https://github.com/container-storage-interface)。在监控，记录等方面，PouchContainer发挥开放作用来接近原生云。