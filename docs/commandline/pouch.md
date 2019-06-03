## pouch

An efficient container engine

### Synopsis

pouch is a client side tool pouch to interact with daemon side process pouchd. Flags and arguments can be input to do what actually you wish. Then pouch parses the flags and arguments and sends a RESTful request to daemon side pouchd.

### Options

```
  -D, --debug              Switch client log level to DEBUG mode
  -h, --help               help for pouch
  -H, --host string        Specify connecting address of Pouch CLI (default "unix:///var/run/pouchd.sock")
      --tlscacert string   Specify CA file of TLS
      --tlscert string     Specify cert file of TLS
      --tlskey string      Specify key file of TLS
      --tlsverify          Use TLS and verify remote
```

### SEE ALSO

* [pouch build](pouch_build.md)	 - Build an image from a Dockerfile
* [pouch checkpoint](pouch_checkpoint.md)	 - Manage checkpoint commands
* [pouch commit](pouch_commit.md)	 - Commit an image from a container
* [pouch cp](pouch_cp.md)	 - Copy files/folders between a container and the local filesystem
* [pouch create](pouch_create.md)	 - Create a new container with specified image
* [pouch events](pouch_events.md)	 - Get real time events from the daemon
* [pouch exec](pouch_exec.md)	 - Run a command in a running container
* [pouch gen-doc](pouch_gen-doc.md)	 - Generate docs
* [pouch history](pouch_history.md)	 - Display history information on image
* [pouch image](pouch_image.md)	 - Manage image
* [pouch images](pouch_images.md)	 - List all images
* [pouch info](pouch_info.md)	 - Display system-wide information
* [pouch inspect](pouch_inspect.md)	 - Get the detailed information of container
* [pouch load](pouch_load.md)	 - load a set of images from a tar archive or STDIN
* [pouch login](pouch_login.md)	 - Login to a registry
* [pouch logout](pouch_logout.md)	 - Logout from a registry
* [pouch logs](pouch_logs.md)	 - Print a container's logs
* [pouch network](pouch_network.md)	 - Manage pouch networks
* [pouch pause](pouch_pause.md)	 - Pause one or more running containers
* [pouch ps](pouch_ps.md)	 - List containers
* [pouch pull](pouch_pull.md)	 - Pull an image from registry
* [pouch push](pouch_push.md)	 - Push an image to registry
* [pouch remount-lxcfs](pouch_remount-lxcfs.md)	 - remount lxcfs bind in containers
* [pouch rename](pouch_rename.md)	 - Rename a container with newName
* [pouch restart](pouch_restart.md)	 - restart one or more containers
* [pouch rm](pouch_rm.md)	 - Remove one or more containers
* [pouch rmi](pouch_rmi.md)	 - Remove one or more images by reference
* [pouch run](pouch_run.md)	 - Create a new container and start it
* [pouch save](pouch_save.md)	 - Save an image to a tar archive or STDOUT
* [pouch search](pouch_search.md)	 - Search the images from specific registry
* [pouch start](pouch_start.md)	 - Start one or more created or stopped containers
* [pouch stats](pouch_stats.md)	 - Display a live stream of container(s) resource usage statistics
* [pouch stop](pouch_stop.md)	 - Stop one or more running containers
* [pouch tag](pouch_tag.md)	 - Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE
* [pouch top](pouch_top.md)	 - Display the running processes of a container
* [pouch unpause](pouch_unpause.md)	 - Unpause one or more paused container
* [pouch update](pouch_update.md)	 - Update the configurations of a container
* [pouch updatedaemon](pouch_updatedaemon.md)	 - Update the configurations of pouchd
* [pouch upgrade](pouch_upgrade.md)	 - Upgrade a container with new image and args
* [pouch version](pouch_version.md)	 - Print versions about Pouch CLI and Pouchd
* [pouch volume](pouch_volume.md)	 - Manage pouch volumes
* [pouch wait](pouch_wait.md)	 - Block until one or more containers stop, then print their exit codes

