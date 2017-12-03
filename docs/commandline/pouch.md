# pouch

pouch is a client side tool `pouch` to interact with pouchd.

## Description

You can use client side tool `pouch` to interact with daemon side process `pouchd`. Flags and arguments can be input to do what actually you wish. Then pouch parses the flags and arguments and sends a RESTful request to daemon side `pouchd`.

## Help Information

``` markdown
Usage:
  pouch [command]

Available Commands:
  create      Create a new container with specified image
  help        Help about any command
  images      show images
  pull        Pull use to download image from repository
  start       Start a created container
  stop        Stop a running container
  version     Print version
  volume      Manage pouch volumes

Flags:
  -h, --help               help for pouch
  -H, --host string        Specify listen address of pouchd (default "unix:///var/run/pouchd.sock")
      --timeout duration   Set timeout (default 10s)
      --tlscacert string   Specify CA file of tls
      --tlscert string     Specify cert file of tls
      --tlskey string      Specify key file of tls
      --tlsverify          Switch if verify the remote when using tls

Use "pouch [command] --help" for more information about a command.
```

## Subcommand Guidance

### create

### helo

### images

### pull

### start

### stop

### version

### volume

## Global Flags

### --help

### --host

### --timeout

### --tlscacert

### --tlscert

### --tlskey

### --tlsverify
