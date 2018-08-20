# PouchContainer with Dragonfly

Container technology is helpful to facilitate IT operation and maintenance, but the distribution of images has brought a huge challenge at the same time. The images can be as large as several Gibs, pulling images will be very slow, not to mention the situation lots of pulling requests simultaneously or low network bandwidth. Dragonfly can play an important role here, as a P2P technology, it provides very efficient distribution, avoiding images distribution as a bottleneck in container technology.

## What is Dragonfly

[Dragonfly](https://github.com/alibaba/Dragonfly#installation) is a P2P file distribution system. It solves the problems of time-consuming distribution, low success rate, and wasteful bandwidth in large-scale file distribution scenarios. Dragonfly significantly improves service capabilities such as data preheating, and large-scale container image distribution.

### composition

#### server

Server consists of cluster managers, it download files from source and constructs a P2P network.

#### client

Client concludes two part: dfget and proxy. dfget is a host-side tool for downloading files, proxy is used for intercepted the http requests and route them to dfget.

## Install dragonfly

1.install server

server can be deployed in two way, on the physical machine or on a container, the steps is simple in the [install dragonfly server](https://github.com/alibaba/Dragonfly/blob/master/docs/install_server.md)

2.install client

- Downloading [client package](https://github.com/alibaba/Dragonfly/blob/master/package/df-client.linux-amd64.tar.gz).
- untar the package `tar xzvf df-client.linux-amd64.tar.gz -C /usr/local/bin`, or you can untar it to any directory you like, but remember to add client path into PATH environment, `export PATH=$PATH:/usr/local/bin/df-client`.
- add server ip into client config file, `/etc/dragonfly.conf`, nodeIp is the host ip which you deploy server

```
[node]
address=nodeIp1,nodeIp2,...
```

about detail install information, you can find in [install dragonfly client](https://github.com/alibaba/Dragonfly/blob/master/docs/install_client.md)

## Run PouchContainer with dragonfly

1.start dragonfly proxy

Start dragonfly proxy, `/usr/local/bin/df-client/df-daemon  --registry https://reg.docker.alibaba-inc.com`, you can add flag `--verbose` to get debug log. Dragonfly logs can be found in
directory `~/.small-dragonfly/logs`.

More dragonfly usage information you can find in [dragonfly usage](https://github.com/alibaba/Dragonfly/blob/master/docs/usage.md)

2.add the following configuration in PouchContainer config file `/etc/pouch/config.json`

```
{
    "image-proxy": "http://127.0.0.1:65001"
}
```

3.pull a image `reg.docker.alibaba-inc.com/base/busybox:latest`, you can find the following output in `~/.small-dragonfly/logs/dfdaemon.log`, which means dragonfly works.

```
time="2018-03-06 20:08:00" level=debug msg="pre access:http://storage.docker.aliyun-inc.com/docker/registry/v2/blobs/sha256/1b/1b5110ff48b0aa7112544e1666cc7199f812243ded4128f0a1b2be027c7    38bec/data?Expires=1520339335&OSSAccessKeyId=LTAIfYaNrksx0ktL&Signature=fVEYIQzIaXyqIcAhypbmzaUx5x8%3D"
time="2018-03-06 20:08:00" level=debug msg="post access:http://storage.docker.aliyun-inc.com"
time="2018-03-06 20:08:00" level=debug msg="pre access:http://storage.docker.aliyun-inc.com/docker/registry/v2/blobs/sha256/98/9869810a78644038c8d37a5a3344de0217cb37bcc2caa2313036c6948b0    ee8da/data?Expires=1520339335&OSSAccessKeyId=LTAIfYaNrksx0ktL&Signature=Tx7JYU07Gap8RfasvCe0JGAUCo4%3D"
```
