# PouchContainer with cri verification

[cri-tools](https://github.com/kubernetes-incubator/cri-tools) provides a CLI([crictl](https://github.com/kubernetes-incubator/cri-tools/blob/master/docs/crictl.md)) for CRI-compatible container runtimes. This is an easy way to verify CRI implementation in PouchContainer without setting up all Kubernetes components.

## Install

The CRI CLI can be installed easily via `go get` command:

```bash
go get github.com/kubernetes-incubator/cri-tools/cmd/crictl
```

Then `crictl` binary can be found in `$GOPATH/bin`.

*Note: ensure GO is installed and GOPATH is set before installing crictl.*

## Usage

```bash
crictl SUBCOMMAND [FLAGS]
```

Subcommands includes(crictl v1.12):

- `attach`:       Attach to a running container
- `create`:       Create a new container
- `exec`:         Run a command in a running container
- `version`:      Display runtime version information
- `images`:       List images
- `inspect`:      Display the status of one or more containers
- `inspecti`:     Return the status of one ore more images
- `inspectp`:     Display the status of one or more pods
- `logs`:         Fetch the logs of a container
- `port-forward`: Forward local port to a pod
- `ps`:           List containers
- `pull`:         Pull an image from a registry
- `runp`:         Run a new pod
- `rm`:           Remove one or more containers
- `rmi`:          Remove one or more images
- `rmp`:          Remove one or more pods
- `pods`:         List pods
- `start`:        Start one or more created containers
- `info`:         Display information of the container runtime
- `stop`:         Stop one or more running containers
- `stopp`:        Stop one or more running pods
- `update`:       Update one or more running containers
- `config`:       Get and set crictl options
- `stats`:        List container(s) resource usage statistics
- `completion`:   Output bash shell completion code
- `help, h`:      Shows a list of commands or help for one command

## Config runtime endpoint

crictl connects to /var/run/dockershim.sock by default. For other runtimes, the endpoint can be set in three ways:

- By setting flags `--runtime-endpoint` and `--image-endpoint`
- By setting environment variables `CRI_RUNTIME_ENDPOINT` and `CRI_IMAGE_ENDPOINT`
- By setting the endpoint in the config file `--config=/etc/crictl.yaml`

```
# cat /etc/crictl.yaml
runtime-endpoint: unix:///var/run/pouchcri.sock
image-endpoint: unix:///var/run/pouchcri.sock
timeout: 10
debug: true
```

## Examples

### Run sandbox with config file

```
# pouch pull k8s.gcr.io/pause-amd64:3.0

# cat sandbox-config.json
{
    "metadata": {
        "name": "nginx-sandbox",
        "namespace": "default",
        "attempt": 1,
        "uid": "hdishd83djaidwnduwk28bcsb"
    },
    "linux": {
    }
}

# crictl runs sandbox-config.json
53bfc944e2e6b391089d441d364e9fea98ea4a51c882d831f5a83d5fd0803162

# pouch ps
Name                                                        ID       Status    Image                              Runtime   Created
k8s_POD_nginx-sandbox_default_hdishd83djaidwnduwk28bcsb_1   53bfc9   running   k8s.gcr.io/pause-amd64:3.0         runc      4 seconds ago
```

### Pull image

```
# crictl pull docker.io/library/redis:alpine
Image is update to date for 0153c5db97e5

# crictl images
IMAGE                        TAG                 IMAGE ID            SIZE
docker.io/library/redis      alpine              0153c5db97e58       10.1MB
```

### unsuccessful cases

If PouchContainer has not fully or correctly implemented some interfaces in CRI, crictl command execution would fail:

```

# crictl ps
FATA[0000] listing containers failed: rpc error: code = Unknown desc = ListContainers Not Implemented Yet
```

