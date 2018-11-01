#!/bin/bash

# This file allows you to init kuberenete master node with pouch containers available.
# Ubuntu and CentOS supported.
# This file is not responsible for CNI network configuration, you need to do it by yourself.

set -o errexit
#set -o nounset

# users can set environments to enable silent installation.
# for example
# export KUBERNETES_VERSION=1.10 CRI_VERSION=v1alpha2 RELEASE_UBUNTU=v1.10.2 MASTER_NODE=true INSTALL_FLANNEL=true INSTALL_SAMPLE=true

if [ -z "$KUBERNETES_VERSION" ] || [ -z "$CRI_VERSION" ]; then
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
        CRI_VERSION="v1alpha1"
        KUBERNETES_RELEASE="1.9.4"
        RELEASE_UBUNTU="1.9.4-00"
        RELEASE_CENTOS="1.9.4-0.x86_64";;
        2)
        KUBERNETES_VERSION="1.10"
        CRI_VERSION="v1alpha2"
        KUBERNETES_RELEASE="1.10.2"
        RELEASE_UBUNTU="1.10.2-00"
        RELEASE_CENTOS="1.10.2-0.x86_64";;
        0)
        exit;;
    esac
fi

echo "KUBERNETES_VERSION:" $KUBERNETES_VERSION

if [ -z "$MASTER_NODE" ]; then
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
fi
echo "MASTER_NODE:"$MASTER_NODE

if [ -z "$CRI_VERSION" ]; then
    echo "CRI_VERSION can't be null" >&2
    exit 1
fi

MASTER_CIDR="10.244.0.0/16"

install_pouch_ubuntu() {
    apt-get -y install lxcfs
    apt-get -y install curl apt-transport-https ca-certificates software-properties-common
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
    mkdir -p /etc/systemd/system/pouch.service.d
    cat > /etc/systemd/system/pouch.service.d/pouch.conf <<EOF
[Service]
ExecStart=
ExecStart=/usr/bin/pouchd --enable-cri=true --cri-version=$CRI_VERSION
EOF
    systemctl daemon-reload
    systemctl restart pouch
}

config_pouch_centos() {
    sed -i "s/ExecStart=\/usr\/local\/bin\/pouchd/ExecStart=\/usr\/local\/bin\/pouchd --enable-cri=true --cri-version=$CRI_VERSION/g" /lib/systemd/system/pouch.service
    systemctl daemon-reload
    systemctl restart pouch
}

config_repo_ubuntu() {
    curl https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -
    cat <<EOF > /etc/apt/sources.list.d/kubernetes.list
deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
EOF
}

config_repo_centos(){
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
    if [ -z "$RELEASE_UBUNTU" ]; then
        echo "RELEASE_UBUNTU can't be null" >&2
        exit 1
    fi

    apt-get update
    apt-get -y install kubelet=$RELEASE_UBUNTU kubeadm=$RELEASE_UBUNTU kubectl=$RELEASE_UBUNTU
}

install_kubelet_centos() {
    if [ -z "$RELEASE_CENTOS" ]; then
        echo "RELEASE_CENTOS can't be null" >&2
        exit 1
    fi

    yum -y install kubelet-$RELEASE_CENTOS kubeadm-$RELEASE_CENTOS kubectl-$RELEASE_CENTOS
    systemctl disable firewalld && systemctl stop firewalld
    systemctl enable kubelet
}

install_cni_ubuntu() {
    apt-get -y install kubernetes-cni
}

install_cni_centos() {
    setenforce 0
    yum install -y kubernetes-cni
}

kubelet_config() {
    sed -i '2 i\Environment="KUBELET_EXTRA_ARGS=--container-runtime=remote --container-runtime-endpoint=unix:///var/run/pouchcri.sock --image-service-endpoint=unix:///var/run/pouchcri.sock"' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
    systemctl daemon-reload
    systemctl enable kubelet
    systemctl start kubelet
}

setup_imagerepository() {
   cat <<EOF > kubeadm.conf
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
kubernetesVersion: $KUBERNETES_RELEASE
api:
  bindPort: 6443
certificatesDir: /etc/kubernetes/pki
clusterName: pouch
imageRepository: registry.cn-hangzhou.aliyuncs.com/google_containers
networking:
  dnsDomain: cluster.local
  serviceSubnet: 10.96.0.0/12
  podSubnet: $MASTER_CIDR
nodeRegistration:
  criSocket: /var/run/pouchcri.sock
EOF
}

setup_master() {
    kubeadm init --config kubeadm.conf --ignore-preflight-errors=all

    # Config default kubeconfig for kubectl
    mkdir -p "${HOME}/.kube"
    cat /etc/kubernetes/admin.conf > "${HOME}/.kube/config"
    chown "$(id -u):$(id -g)" "${HOME}/.kube/config"

    until kubectl get nodes &> /dev/null; do echo "Waiting kubernetes api server for a second..."; sleep 1; done
    # Enable master node scheduling
    kubectl taint nodes --all  node-role.kubernetes.io/master-
    if $INSTALL_FLANNEL; then
        install_flannel
    fi
}

# Install flannel for the cluster
install_flannel(){
    kubectl apply -f https://github.com/coreos/flannel/raw/master/Documentation/kube-flannel.yml
}

# create a sample
create_sample(){
    kubectl apply -f - <<EOF
---
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
        version: v1
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 80
          protocol: TCP
        args:
        env:
          - name: OWNER
            value: "Pouch"
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 30
          timeoutSeconds: 30
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 30
          timeoutSeconds: 30
---
kind: Service
apiVersion: v1
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: nginx
EOF
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
        config_repo_ubuntu
        install_kubelet_ubuntu
        install_cni_ubuntu
        kubelet_config
        setup_imagerepository
        if $MASTER_NODE; then
            setup_master
            if $INSTALL_SAMPLE; then
                create_sample
            fi
        fi
    ;;

    fedora|centos|redhat)
        install_pouch_centos
        config_pouch_centos
        config_repo_centos
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