# create

Command `create` is used to create a container according to user's configuration.

## Description

The `pouch create` command creates a container object in Pouchd. When creating, all configuration user input will be stored in memory store of Pouchd. This is useful when you wish to create a container configuration ahead of time so that Pouchd will preserve the resource in advance. The container you created is ready to start when you need it.

## Help Information

``` shell
$ pouch create --help
Create a new container with specified image

Usage:
  pouch create [image] [flags]

Flags:
  -h, --help             help for create
      --name string      specified the container's name
  -t, --tty              allocate a tty device
  -v, --volume strings   create container with volumes

Global Flags:
  -H, --host string        Specify listen address of pouchd (default "unix:///var/run/pouchd.sock")
      --tlscacert string   Specify CA file of tls
      --tlscert string     Specify cert file of tls
      --tlskey string      Specify key file of tls
      --tlsverify          Switch if verify the remote when using tls
```

## Flag Guidance

### --name

User can set container's name by passing flag `--name`. For example:

``` shell
$ pouch create --name foo busybox:latest
container ID: e1d541722d68dc5d133cca9e7bd8fd9338603e1763096c8e853522b60d11f7b9, name: foo
```

### -t, --tty

Allocate a pseudo-TTY for container.

### -v, --volume

Bind created volumes or host path to container.

### -h, --help

User can get help information via `-h` flag.

### Global Flags

All CLI side commands have the same global flags. For more details, please refer to [Global Flags](pouch.md#global-flags).