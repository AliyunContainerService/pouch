#!/bin/bash

# This file allows you to init kuberenete master node with pouch containers available.
# Ubuntu and CentOS supported.
# This file is not responsible for CNI network configuration, you need to do it by yourself.

set -o errexit
set -o nounset

echo "----------------------------------"
echo "Please choose the version of kubernetes:"
echo "(1) kubernetes 1.9"
echo "(2) kubernetes 1.10"
echo "(0) exit"
echo "----------------------------------"
read input
case $input in
    1)
    KUBERNETES_VERSION="1.9"
    KUBERNETES_VERSION_UBUNTU="1.9.4-00"
    KUBERNETES_VERSION_CENTOS="1.9.4"
    CRI_VERSION="v1alpha1"
    RELEASE_UBUNTU="v1.9.4"
    RELEASE_CENTOS="1.9.4-0.x86_64";;
    2)
    KUBERNETES_VERSION="1.10"
    KUBERNETES_VERSION_UBUNTU="1.10.2-00"
    KUBERNETES_VERSION_CENTOS="1.10.2"
    CRI_VERSION="v1alpha2"
    RELEASE_UBUNTU="v1.10.2"
    RELEASE_CENTOS="1.10.2-0.x86_64";;
    0)
    exit;;
esac
echo "KUBERNETES_VERSION:" $KUBERNETES_VERSION

echo "----------------------------------"
echo "Is it the master node?"
echo "(Y/y) Y"
echo "(N/n) N"
echo "(0) exit"
echo "----------------------------------"
read input
case $input in
    Y | y )
    MASTER_NODE="true";;
    N | n)
    MASTER_NODE="false";;
    0)
    exit;;
esac
echo "MASTER_NODE:"$MASTER_NODE

MASTER_CIDR="10.244.0.0/16"
CNI_VERSION="v0.6.0"

install_pouch_ubuntu() {
    apt-get install lxcfs
    apt-get install curl apt-transport-https ca-certificates software-properties-common
    curl -fsSL http://mirrors.aliyun.com/opsx/pouch/linux/debian/opsx@service.alibaba.com.gpg.key | sudo apt-key add -
    add-apt-repository "deb http://mirrors.aliyun.com/opsx/pouch/linux/debian/ pouch stable"
    apt-get -y update
    apt-get install -y pouch
    systemctl enable pouch
    systemctl start pouch
}

install_pouch_centos() {
    yum install -y yum-utils
    yum-config-manager --add-repo http://mirrors.aliyun.com/opsx/opsx-centos7.repo
    yum -y update
    yum install -y pouch
    systemctl enable pouch
    systemctl start pouch
}

config_pouch_ubuntu() {
    sed -i "s/ExecStart=\/usr\/bin\/pouchd/ExecStart=\/usr\/bin\/pouchd --enable-cri=true --cri-version=$CRI_VERSION/g" /usr/lib/systemd/system/pouch.service
    systemctl daemon-reload
    systemctl restart pouch
}

config_pouch_centos() {
    sed -i "s/ExecStart=\/usr\/local\/bin\/pouchd/ExecStart=\/usr\/local\/bin\/pouchd --enable-cri=true --cri-version=$CRI_VERSION/g" /lib/systemd/system/pouch.service
    systemctl daemon-reload
    systemctl restart pouch
}

config_repo(){
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
}

install_kubelet_ubuntu() {
    KUBE_URL="https://storage.googleapis.com/kubernetes-release/release/$RELEASE_UBUNTU/bin/linux/amd64"
    wget "$KUBE_URL/kubeadm" -O /usr/bin/kubeadm
    wget "$KUBE_URL/kubelet" -O /usr/bin/kubelet
    wget "$KUBE_URL/kubectl" -O /usr/bin/kubectl
    chmod +x /usr/bin/kubeadm /usr/bin/kubelet /usr/bin/kubectl

    KUBELET_URL="https://raw.githubusercontent.com/kubernetes/kubernetes/$RELEASE_UBUNTU/build/debs"
    mkdir -p /etc/systemd/system/kubelet.service.d
    wget "$KUBELET_URL/kubelet.service" -O /etc/systemd/system/kubelet.service
    wget "$KUBELET_URL/10-kubeadm.conf" -O /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
}

install_kubelet_centos() {
    yum -y install kubelet-$RELEASE_CENTOS kubeadm-$RELEASE_CENTOS kubectl-$RELEASE_CENTOS
    systemctl disable firewalld && systemctl stop firewalld
    systemctl enable kubelet
}

install_cni_ubuntu() {
    mkdir -p /opt/cni/bin
    curl -L "https://github.com/containernetworking/plugins/releases/download/$CNI_VERSION/cni-plugins-amd64-$CNI_VERSION.tgz" | tar -C /opt/cni/bin -xz
}

install_cni_centos() {
    setenforce 0
    yum install -y kubernetes-cni
}

kubelet_config() {
    sed -i '2 i\Environment="KUBELET_EXTRA_ARGS=--container-runtime=remote --container-runtime-endpoint=unix:///var/run/pouchcri.sock --image-service-endpoint=unix:///var/run/pouchcri.sock"' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
    systemctl daemon-reload
    systemctl start kubelet
}

setup_imagerepository() {
   cat <<EOF > kubeadm.conf
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
imageRepository: registry.cn-hangzhou.aliyuncs.com/google_containers
kubernetes-version: stable-$KUBERNETES_VERSION
EOF
}

setup_master() {
    kubeadm init --config kubeadm.conf --ignore-preflight-errors=all
}

command_exists() {
    command -v "$@" > /dev/null 2>&1
}

lsb_dist=''
if command_exists lsb_release; then
    lsb_dist="$(lsb_release -si)"
fi
if [ -z "$lsb_dist" ] && [ -r /etc/lsb-release ]; then
    lsb_dist="$(. /etc/lsb-release && echo "$DISTRIB_ID")"
fi
if [ -z "$lsb_dist" ] && [ -r /etc/centos-release ]; then
    lsb_dist='centos'
fi
if [ -z "$lsb_dist" ] && [ -r /etc/redhat-release ]; then
    lsb_dist='redhat'
fi
if [ -z "$lsb_dist" ] && [ -r /etc/os-release ]; then
    lsb_dist="$(. /etc/os-release && echo "$ID")"
fi

lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"

case "$lsb_dist" in

    ubuntu)
        install_pouch_ubuntu
        config_pouch_ubuntu
        install_kubelet_ubuntu
        install_cni_ubuntu
        kubelet_config
        setup_imagerepository
        if $MASTER_NODE; then
            setup_master
        fi
    ;;

    fedora|centos|redhat)
        install_pouch_centos
        config_pouch_centos
        config_repo
        install_kubelet_centos
        kubelet_config
        install_cni_centos
        setup_imagerepository
        if $MASTER_NODE; then
            setup_master
        fi
    ;;

    *)
        echo "$lsb_dist is not supported (not in centos|ubuntu)"
    ;;

esac