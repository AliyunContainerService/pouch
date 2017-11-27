ctrd

ctrd is a package to interact with containerd using containerd's exposed API. Pouch will vendor package containerd in repo [containerd/containerd](https://github.com/containerd/containerd) and [alibaba/containerd](https://github.com/containerd/containerd).

## difference between containerds

First, we should said alibaba/containerd is definitely forked from upstream project containerd/containerd.

containerd/containerd definitely has the code base of the other one. Then what is the usage of alibaba/containerd. We say it is **BACK PORTING**.

Every time releasing Pouch to launch a new version, we vendor containerd/containerd. Since upstream containerd/containerd keeps its own way to evolve, when a serious bug is fixed in the upstream and Pouch needs this for backporting only, it is not proper to vendor the latest container/containerd which contains too many changes. As a result, we come up with a way to backport severe bugfix in alibaba/containerd. Then when backporting, we vendor alibaba/containerd, otherwise we still take advantages of upstream containerd/containerd.

To check if the package we use is containerd/containerd or alibaba/containerd, we can refer to [vendor.json](../vendor/vendor.json).

 
