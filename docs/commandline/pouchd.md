## pouchd

An Efficient Enterprise-class Container Engine

### Synopsis

An Efficient Enterprise-class Container Engine

```
pouchd [flags]
```

### Options

```
      --add-runtime runtime                 register a OCI runtime to daemon (default [])
      --allow-multi-snapshotter             If set true, pouchd will allow multi snapshotter
      --bip string                          Set bridge IP
      --bridge-name string                  Set default bridge name
      --cgroup-parent string                Set parent cgroup for all containers (default "default")
      --cni-bin-dir string                  The directory for putting cni plugin binaries. (default "/opt/cni/bin")
      --cni-conf-dir string                 The directory for putting cni plugin configuration files. (default "/etc/cni/net.d")
      --config-file string                  Configuration file of pouchd (default "/etc/pouch/config.json")
  -c, --containerd string                   Specify listening address of containerd (default "/var/run/containerd.sock")
      --containerd-path string              Specify the path of containerd binary
      --cri-stats-collect-period int        The time duration (in time.Second) cri collect stats from containerd. (default 10)
      --cri-version string                  Specify the version of cri which is used to support Kubernetes (default "v1alpha2")
  -D, --debug                               Switch daemon log level to DEBUG mode
      --default-gateway string              Set default IPv4 bridge gateway
      --default-gateway-v6 string           Set default IPv6 bridge gateway
      --default-namespace string            default-namespace is passed to containerd, the default value is 'default' (default "default")
      --default-registry string             Default Image Registry (default "registry.hub.docker.com")
      --default-registry-namespace string   Default Image Registry namespace (default "library")
      --default-runtime string              Default OCI Runtime (default "runc")
      --disable-cri-stats-collect           Specify whether cri collect stats from containerd.If this is true, option CriStatsCollectPeriod will take no effect. (default true)
      --enable-cri                          Specify whether enable the cri part of pouchd which is used to support Kubernetes
      --enable-ipv6                         Enable IPv6 networking
      --enable-lxcfs                        Enable Lxcfs to make container to isolate /proc
      --enable-profiler                     Set if pouchd setup profiler
      --exec-root-dir string                Set exec root directory for network
      --fixed-cidr string                   Set bridge fixed CIDRv4
      --fixed-cidr-v6 string                Set bridge fixed CIDRv6
  -h, --help                                help for pouchd
      --home-dir string                     Specify root dir of pouchd (default "/var/lib/pouch")
      --image-proxy string                  Http proxy to pull image
      --ipforward                           Enable ipforward (default true)
      --iptables                            Enable iptables (default true)
      --label strings                       Set metadata for Pouch daemon
  -l, --listen stringArray                  Specify listening addresses of Pouchd (default [unix:///var/run/pouchd.sock])
      --listen-cri string                   Specify listening address of CRI (default "unix:///var/run/pouchcri.sock")
      --log-driver string                   Set default log driver (default "json-file")
      --log-opt stringArray                 Set default log driver options
      --lxcfs string                        Specify the path of lxcfs binary (default "/usr/local/bin/lxcfs")
      --lxcfs-home string                   Specify the mount dir of lxcfs (default "/var/lib/lxcfs")
      --manager-whitelist string            Set tls name whitelist, multiple values are separated by commas
      --mtu int                             Set bridge MTU (default 1500)
      --oom-score-adj int                   Set the oom_score_adj for the daemon (default -500)
      --pidfile string                      Save daemon pid (default "/var/run/pouch.pid")
      --quota-driver string                 Set quota driver(grpquota/prjquota), if not set, it will set by kernel version
      --sandbox-image string                The image used by sandbox container. (default "registry.cn-hangzhou.aliyuncs.com/google-containers/pause-amd64:3.0")
      --snapshotter string                  Snapshotter driver of pouchd, it will be passed to containerd (default "overlayfs")
      --stream-server-port string           The port stream server of cri is listening on. (default "10010")
      --stream-server-reuse-port            Specify whether cri stream server share port with pouchd. If this is true, the listen option of pouchd should specify a tcp socket and its port should be same with stream-server-port.
      --tlscacert string                    Specify CA file of TLS
      --tlscert string                      Specify cert file of TLS
      --tlskey string                       Specify key file of TLS
      --tlsverify                           Use TLS and verify remote
      --userland-proxy                      Enable userland proxy
  -v, --version                             Print daemon version
      --volume-driver-alias string          Set volume driver alias, <name=alias>[;name1=alias1]
```

### SEE ALSO

* [pouchd gen-doc](pouchd_gen-doc.md)	 - Generate document for pouchd CLI with MarkDown format