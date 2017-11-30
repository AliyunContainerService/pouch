# pouchd

`pouchd` runs a long-running process managing containers on host.

## Description

Users can execute `pouchd` to run long running process which manages images, containers, volumes and so on. The running pouchd process can accept requests from pouch cli, handle requests and manage containers. `pouchd` is a long-running process background, and you can config for it by passing command line flags which is defined in pouchd.

## Help Information

For more help information on pouchd's runtime, you can get details via `pouchd --help`

``` markdown
Usage:
   [flags]

Flags:
  -c, --containerd string          where does containerd listened on (default "/var/run/containerd.sock")
      --containerd-config string   Specify the path of Containerd binary (default "/etc/containerd/config.toml")
      --containerd-path string     Specify the path of Containerd binary (default "/usr/local/bin/containerd")
  -D, --debug                      switch debug level
  -h, --help                       help for this command
      --home-dir string            The pouchd's home directory (default "/var/lib/pouch")
  -l, --listen stringArray         which address to listen on (default [unix:///var/run/pouchd.sock])
      --tlscacert string           Specify CA file of tls
      --tlscert string             Specify cert file of tls
      --tlskey string              Specify key file of tls
      --tlsverify                  Switch if verify the remote when using tls
``` 

## Flag Guidance

`dockerd` has lots of flags to config how to run pouch daemon. This flags cover multi fields of pouch, such as security, storage, network and so on. The follwing content includes detailed illustration of each flag.

### --containerd

### --containerd-config

### --containerd-path

### --debug

### --help

### --home-dir

### --listen

### --tlscacert

### --tlscert

### --tlskey

### --tlsverify


