# Deploy Kubernetes With PouchContainer, Powered By Aliyun

Updated: 2018.6.1

- [PouchContainer deploying](#pouch-with-kubernetes-deploying)
  - [Overview](#overview)
  - [Restriction](#restriction)
  - [Install and Configure](#install-and-configure)
    - [Install PouchContainer](#install-pouch)
    - [Setup Repo](#setup-repo)
    - [Install Kubernetes Components](#install-kubernetes-components)
    - [Install CNI](#install-cni)
    - [Setting up the master node](#setting-up-the-master-node)
    - [Setting up ImageRepository](#setting-up-imagerepository)
    - [Setting up the minion nodes](#setting-up-the-minion-nodes)
  - [Run and Verify](#run-and-verify)
  - [Troubleshooting](#troubleshooting)

## Overview

This document shows how to easily install a Kubernetes cluster with PouchContainer as the container runtime using Aliyun Image Repository.

![pouch_with_kubernetes](../static_files/pouch_with_kubernetes.png)

## Restriction

Kubernetes: Version 1.5+ is recommanded.

NOTE: PouchContainer version prior to 0.5.x (including version 0.5.0) did not support configuring  CNI network plugin with flannel. If you want to do that, use the latest code from the branch of master, refer to  [Developer Quick-Start](https://github.com/alibaba/pouch/blob/master/INSTALLATION.md#developer-quick-start)

## Install and Configure

An all-in-one kubernetes cluster with PouchContainer runtime could be deployed by running:

```
hack/kubernetes/allinone_aliyun.sh
```

Please refer to [allinone_aliyun](https://github.com/alibaba/pouch/blob/master/hack/kubernetes/allinone_aliyun.sh) .

### Install PouchContainer

You can easily setup a basic PouchContainer environment, see [INSTALLATION.md](../../INSTALLATION.md).

### Configure PouchContainer

On Ubuntu 16.04+:

NOTE: If you'd like to use Kubernetes 1.10+, CRI_VERSION should be "v1alpha2"

```
CRI_VERSION="v1alpha1"
sed -i 's/ExecStart=\/usr\/bin\/pouchd/ExecStart=\/usr\/bin\/pouchd --enable-cri=true --cri-version=${CRI_VERSION}/g' /usr/lib/systemd/system/pouch.service
systemctl daemon-reload
systemctl restart pouch
```

On CentOS 7:

NOTE: If you'd like to use Kubernetes 1.10+, CRI_VERSION should be "v1alpha2"

```
CRI_VERSION="v1alpha1"
sed -i 's/ExecStart=\/usr\/local\/bin\/pouchd/ExecStart=\/usr\/local\/bin\/pouchd --enable-cri=true --cri-version=${CRI_VERSION}/g' /lib/systemd/system/pouch.service
systemctl daemon-reload
systemctl restart pouch
```

### Setup Repo

On CentOS 7:

```
cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=http://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=0
repo_gpgcheck=0
gpgkey=http://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg
       http://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF
```

### Install Kubernetes Components

On Ubuntu 16.04+:

```sh
RELEASE="v1.9.4"
KUBE_URL="https://storage.googleapis.com/kubernetes-release/release/${RELEASE_UBUNTU}/bin/linux/amd64"
wget "${KUBE_URL}/kubeadm" -O /usr/bin/kubeadm
wget "${KUBE_URL}/kubelet" -O /usr/bin/kubelet
wget "${KUBE_URL}/kubectl" -O /usr/bin/kubectl
chmod +x /usr/bin/kubeadm /usr/bin/kubelet /usr/bin/kubectl

KUBELET_URL="https://raw.githubusercontent.com/kubernetes/kubernetes/${RELEASE_UBUNTU}/build/debs"
mkdir -p /etc/systemd/system/kubelet.service.d
wget "${KUBELET_URL}/kubelet.service" -O /etc/systemd/system/kubelet.service
wget "${KUBELET_URL}/10-kubeadm.conf" -O /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
```

On CentOS 7:

```sh
RELEASE="1.9.4-0.x86_64"
yum -y install kubelet-${RELEASE} kubeadm-${RELEASE} kubectl-${RELEASE}
```

Configure kubelet with PouchContainer as its runtime:

```sh
sed -i '2 i\Environment="KUBELET_EXTRA_ARGS=--container-runtime=remote --container-runtime-endpoint=unix:///var/run/pouchcri.sock --image-service-endpoint=unix:///var/run/pouchcri.sock"' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
systemctl daemon-reload
```

For more details, please check [install kubelet](https://kubernetes.io/docs/setup/independent/install-kubeadm/#installing-kubeadm-kubelet-and-kubectl).

### Install CNI

On Ubuntu 16.04+:

```
CNI_VERSION="v0.6.0"
mkdir -p /opt/cni/bin
curl -L "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-amd64-${CNI_VERSION}.tgz" | tar -C /opt/cni/bin -xz
```

On CentOS 7:

```
setenforce 0
yum install -y kubernetes-cni
```

### Setting up ImageRepository

```
# cat kubeadm.conf
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
imageRepository: registry.cn-hangzhou.aliyuncs.com/google_containers
kubernetes-version: stable-1.9
networking:
  podSubnet: 10.244.0.0/16
```

### Setting up the master node

For more detailed Kubernetes cluster installation, please check [Using kubeadm to Create a Cluster](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/)

```
kubeadm init --config kubeadm.conf --ignore-preflight-errors=all
```

Set the KUBECONFIG environment variable

```sh
export KUBECONFIG=/etc/kubernetes/admin.conf
```

Configure CNI network plugin with [flannel](https://github.com/coreos/flannel)

```
kubectl create -f https://github.com/coreos/flannel/raw/master/Documentation/kube-flannel.yml
```

NOTE: For other plugins, please check [Installing a pod network](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/#pod-network).

Optional: enable schedule pods on the master node

```sh
kubectl taint nodes --all node-role.kubernetes.io/master:NoSchedule-
```

### Setting up the minion nodes

After initializing the master node, you may get the following prompt:

```
You can now join any number of machines by running the following on each node
as root:

  kubeadm join --token $token ${master_ip:port} --discovery-token-ca-cert-hash $ca-cert
```

NOTE: Because kubeadm still assumes docker as the only container runtime ,Use the flag `--ignore-preflight-errors=all` to skip the check.

Copy & Run it in all your minion nodes.

## Run and Verify

Create a deployment named `Pouch`:

```sh
# cat pouch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pouch
  labels:
    pouch: pouch
spec:
  selector:
    matchLabels:
      pouch: pouch
  template:
    metadata:
      labels:
        pouch: pouch
    spec:
      containers:
      - name: pouch
        image: docker.io/library/nginx:latest
        ports:
        - containerPort: 80

# kubectl create -f pouch.yaml
deployment "pouch" created
```

Confirm the pod of deployment is really running:

```sh
# kubectl get pods -o wide
NAME                     READY     STATUS    RESTARTS   AGE       IP           NODE
pouch-7dcd875d69-gq5r9   1/1       Running   0          44m       10.244.1.4   master
# curl 10.244.1.4
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
    body {
        width: 35em;
        margin: 0 auto;
        font-family: Tahoma, Verdana, Arial, sans-serif;
    }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```

## Troubleshooting

- Because `kubeadm` still assumes docker as the only container runtime which can be used with kubernetes. When you use `kubeadm` to initialize the master node or join the minion node to the cluster, you may encounter the following error message:`[ERROR SystemVerification]: failed to get docker info: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?`. Use the flag `--ignore-preflight-errors=all` to skip the check, like `kubeadm init --ignore-preflight-errors=all`.

- By default PouchContainer will support CRI v1alpha2,which means that using a version of Kubernetes prior to 1.10 will not work. As the NOTE mentioned above, we could start pouchd with the configuration like `pouchd --cri-version v1alpha1` to specify the version of CRI to support the version of Kubernetes below 1.10.

- By default PouchContainer will not enable the CRI. If you'd like to deploy Kubernetes with PouchContainer, you should start pouchd with the configuration like `pouchd --enable-cri`.

- By default PouchContainer will use `registry.cn-hangzhou.aliyuncs.com/google-containers/pause-amd64:3.0` as the image of infra container. If you'd like use image other than that, you could start pouchd with the configuration like `pouchd --enable-cri --sandbox-image XXX`.

- Any other troubles? Make an issue to connect with us!
