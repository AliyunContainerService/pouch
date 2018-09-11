# Vagrant support for PouchContainer

You can using [Vagrant](https://www.vagrantup.com) to quickly experience PouchContainer or cross compile on non-linux.

## Requirements

* Vagrant 1.9.x or newer
* VirtuaBox

## Getting Started

```bash
vagrant up
vagrant ssh -c "sudo -i"

# Start a nginx container with 80
pouch run -d --name nginx -p 80:80 nginx
curl http://localhost
```

## Getting Started with Kubernetes

```bash
# On macOS or Linux
POUCH_KUBE=true vagrant up

# On Windows
set POUCH_KUBE=true
vagrant up

vagrant ssh -c "sudo -i"

$ kubectl cluster-info
Kubernetes master is running at https://10.0.2.15:6443
KubeDNS is running at https://10.0.2.15:6443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy


$ kubectl get cs
NAME                 STATUS    MESSAGE
controller-manager   Healthy   ok
scheduler            Healthy   ok
etcd-0               Healthy   {"health": "true"}

$ kubectl get po -o wide --all-namespaces
NAMESPACE     NAME                            READY     STATUS    RESTARTS   AGE       IP           NODE
default       nginx-6dc97b4cbd-cq4pb          1/1       Running   0          2m        10.244.0.2   pouch
default       nginx-6dc97b4cbd-ktlwc          1/1       Running   0          2m        10.244.0.3   pouch
kube-system   etcd-pouch                      1/1       Running   0          3m        10.0.2.15    pouch
kube-system   kube-apiserver-pouch            1/1       Running   0          2m        10.0.2.15    pouch
kube-system   kube-controller-manager-pouch   1/1       Running   0          2m        10.0.2.15    pouch
kube-system   kube-dns-b4bd9576-gwqzv         3/3       Running   0          2m        10.244.0.4   pouch
kube-system   kube-flannel-ds-amd64-vd466     1/1       Running   1          2m        10.0.2.15    pouch
kube-system   kube-proxy-c8l8j                1/1       Running   0          2m        10.0.2.15    pouch
kube-system   kube-scheduler-pouch            1/1       Running   0          2m        10.0.2.15    pouch
```

## Build pouch with vagrant

```bash
# On macOS or Linux
POUCH_BUILD=true vagrant up

# On Windows
set POUCH_BUILD=true
vagrant up

# Install compiled pouch binaries for pouch service.
vagrant ssh -c "sudo -i"
cd ~/go/src/github.com/alibaba/pouch
make PREFIX=/usr install
systemctl restart pouch
pouch version
```
