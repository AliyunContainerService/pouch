## pouch

An efficient container engine

### Synopsis

pouch is a client side tool pouch to interact with daemon side process pouchd. Flags and arguments can be input to do what actually you wish. Then pouch parses the flags and arguments and sends a RESTful request to daemon side pouchd.

### Options

```
  -h, --help               help for pouch
  -H, --host string        Specify connecting address of Pouch CLI (default "unix:///var/run/pouchd.sock")
      --tlscacert string   Specify CA file of TLS
      --tlscert string     Specify cert file of TLS
      --tlskey string      Specify key file of TLS
      --tlsverify          Use TLS and verify remote
```

### SEE ALSO

* [pouch create](pouch_create.md)	 - Create a new container with specified image
* [pouch exec](pouch_exec.md)	 - Exec a process in a running container
* [pouch gen-doc](pouch_gen-doc.md)	 - Generate docs
* [pouch image](pouch_image.md)	 - Manage image
* [pouch images](pouch_images.md)	 - List all images
* [pouch inspect](pouch_inspect.md)	 - Get the detailed information of container
* [pouch login](pouch_login.md)	 - Login to a registry
* [pouch network](pouch_network.md)	 - Manage pouch networks
* [pouch pause](pouch_pause.md)	 - Pause a running container
* [pouch ps](pouch_ps.md)	 - List containers
* [pouch pull](pouch_pull.md)	 - Pull an image from registry
* [pouch rename](pouch_rename.md)	 - Rename a container with newName
* [pouch rm](pouch_rm.md)	 - Remove one or more containers
* [pouch rmi](pouch_rmi.md)	 - Remove one or more images by reference
* [pouch run](pouch_run.md)	 - Create a new container and start it
* [pouch start](pouch_start.md)	 - Start a created or stopped container
* [pouch stop](pouch_stop.md)	 - Stop a running container
* [pouch unpause](pouch_unpause.md)	 - Unpause a paused container
* [pouch update](pouch_update.md)	 - Update the configurations of a container
* [pouch version](pouch_version.md)	 - Print versions about Pouch CLI and Pouchd
* [pouch volume](pouch_volume.md)	 - Manage pouch volumes

