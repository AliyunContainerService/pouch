#!/bin/bash

# This file allows you to init kuberenete master node with pouch containers available.
# Ubuntu and CentOS supported.

set -o errexit
set -o nounset

KUBERNETES_VERSION_UBUNTU="1.9.4-00"
KUBERNETES_VERSION_CENTOS="1.9.4"
MASTER_CIDR="10.244.1.0/24"

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
	sed -i 's/ExecStart=\/usr\/bin\/pouchd/ExecStart=\/usr\/bin\/pouchd --enable-cri=true/g' /usr/lib/systemd/system/pouch.service
	systemctl daemon-reload
	systemctl restart pouch
}

config_pouch_centos() {
	sed -i 's/ExecStart=\/usr\/local\/bin\/pouchd/ExecStart=\/usr\/local\/bin\/pouchd --enable-cri=true/g' /lib/systemd/system/pouch.service
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

config_cni() {
	mkdir -p /etc/cni/net.d
	cat >/etc/cni/net.d/10-mynet.conf <<-EOF
{
    "cniVersion": "0.3.0",
    "name": "mynet",
    "type": "bridge",
    "bridge": "cni0",
    "isGateway": true,
    "ipMasq": true,
    "ipam": {
        "type": "host-local",
        "subnet": "${MASTER_CIDR}",
        "routes": [
            { "dst": "0.0.0.0/0"  }
        ]
    }
}
EOF
	cat >/etc/cni/net.d/99-loopback.conf <<-EOF
{
    "cniVersion": "0.3.0",
    "type": "loopback"
}
EOF
}

setup_master() {
	kubeadm init --skip-preflight-checks
	# enable schedule pods on the master node
	export KUBECONFIG=/etc/kubernetes/admin.conf
	kubectl taint nodes --all node-role.kubernetes.io/master:NoSchedule-
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
        config_cni
        setup_master
    ;;

    fedora|centos|redhat)
        install_pouch_centos
        config_pouch_centos
        install_kubelet_centos
        install_cni_centos
        config_kubelet
        config_cni
        setup_master
    ;;

    *)
        echo "$lsb_dist is not supported (not in centos|ubuntu)"
    ;;

esac
