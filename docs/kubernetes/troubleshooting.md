# Troubleshooting

- Because `kubeadm` still assumes docker as the only container runtime which can be used with kubernetes. When you use `kubeadm` to initialize the master node or join the minion node to the cluster, you may encounter the following error message:`[ERROR SystemVerification]: failed to get docker info: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?`. Use the flag `--ignore-preflight-errors=all` to skip the check, like `kubeadm init --ignore-preflight-errors=all`.

- By default PouchContainer will support CRI v1alpha2,which means that using a version of Kubernetes prior to 1.10 will not work. As the NOTE mentioned above, we could start pouchd with the configuration like `pouchd --cri-version v1alpha1` to specify the version of CRI to support the version of Kubernetes below 1.10.

- By default PouchContainer will not enable the CRI. If you'd like to deploy Kubernetes with PouchContainer, you should start pouchd with the configuration like `pouchd --enable-cri`.

- By default PouchContainer will use `registry.cn-hangzhou.aliyuncs.com/google-containers/pause-amd64:3.0` as the image of infra container. If you'd like use image other than that, you could start pouchd with the configuration like `pouchd --enable-cri --sandbox-image XXX`.

- Any other troubles? Make an issue to connect with us!