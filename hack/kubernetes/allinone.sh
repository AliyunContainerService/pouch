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
    KUBERNETES_VERSION_UBUNTU="1.9.4-00"
    KUBERNETES_VERSION_CENTOS="1.9.4"
    CRI_VERSION="v1alpha1";;
    2)
    KUBERNETES_VERSION_UBUNTU="1.10.2-00"
    KUBERNETES_VERSION_CENTOS="1.10.2"
    CRI_VERSION="v1alpha2";;
    0)
    exit;;
esac
echo "KUBERNETES_VERSION:" $KUBERNETES_VERSION_CENTOS

echo "----------------------------------"
echo "Is it the master node ?"
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

install_pouch_ubuntu() {
	apt-get install lxcfs
	apt-get install curl apt-transport-https ca-certificates software-properties-common
	curl -fsSL http://mirrors.aliyun.com/opsx/pouch/linux/debian/opsx@service.alibaba.com.gpg.key | sudo apt-key add -
	add-apt-repository "deb http://mirrors.aliyun.com/opsx/pouch/linux/debian/ pouch stable"
	apt-get -y update
	apt-get install -y pouch
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
    mkdir -p /etc/systemd/system/pouch.service.d
    cat > /etc/systemd/system/pouch.service.d/pouch.conf <<EOF
[Service]
ExecStart=
ExecStart=/usr/local/bin/pouchd --enable-cri=true --cri-version=$CRI_VERSION
EOF
    systemctl daemon-reload
    systemctl restart pouch
}

config_pouch_centos() {
	sed -i "s/ExecStart=\/usr\/local\/bin\/pouchd/ExecStart=\/usr\/local\/bin\/pouchd --enable-cri=true --cri-version=$CRI_VERSION/g" /lib/systemd/system/pouch.service
    systemctl daemon-reload
	systemctl restart pouch
}

install_kubelet_ubuntu() {
	apt-get update && apt-get install -y apt-transport-https
	curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
	cat <<EOF > /etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF
	apt-get update
	apt-get install -y kubelet=$KUBERNETES_VERSION_UBUNTU kubeadm=$KUBERNETES_VERSION_UBUNTU kubectl=$KUBERNETES_VERSION_UBUNTU
}

install_kubelet_centos() {
	cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=http://yum.kubernetes.io/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
       https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF
	yum install -y kubelet-$KUBERNETES_VERSION_CENTOS kubeadm-$KUBERNETES_VERSION_CENTOS kubectl-$KUBERNETES_VERSION_CENTOS
	systemctl disable firewalld && systemctl stop firewalld
	systemctl enable kubelet
}

install_cni_ubuntu() {
	apt-get install -y kubernetes-cni
}

install_cni_centos() {
	setenforce 0
	yum install -y kubernetes-cni
}

config_kubelet() {
	sed -i '2 i\Environment="KUBELET_EXTRA_ARGS=--container-runtime=remote --container-runtime-endpoint=unix:///var/run/pouchcri.sock --image-service-endpoint=unix:///var/run/pouchcri.sock"' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
	systemctl daemon-reload
	systemctl start kubelet
}

setup_master() {
	kubeadm init --pod-network-cidr $MASTER_CIDR --ignore-preflight-errors=all
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
        config_kubelet
        if $MASTER_NODE; then
            setup_master
        fi
    ;;

    fedora|centos|redhat)
        install_pouch_centos
        config_pouch_centos
        install_kubelet_centos
        install_cni_centos
        config_kubelet
        if $MASTER_NODE; then
            setup_master
        fi
    ;;

    *)
        echo "$lsb_dist is not supported (not in centos|ubuntu)"
    ;;

esac
