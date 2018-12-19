# PouchContainer with plugin

## Why provide plugin

There are many scenes to use the container，the different container engine users, in addition to the general process of using the container operation, perhaps they will have their own customized operation scenarios, pouch-container provides users with a situation that does not affect the general scene. A plug-in mechanism for customizable operations to handle container operations for business privatization.

## Plugin function

The pouch-container plugin provides developers or users with a way to call their own plugin logic, calling plugin logic before or after certain operations to satisfy the specific logic of the plugin implementation.

## Which plugins

Currently pouch-container provides four plugins，they are: `container plugin`，`daemon plugin`，`volume plugin`，`cri plugin`, `image plugin`

### container plugin

In order to run custom code provided by users which will be triggered at some point, we support a plugin framework.

* pre-create container point, at this point you can change the input stream by some rules, in some companies they have some stale orchestration system who use env to pass-in some limit which is an attribute in PouchContainer, then you can use this point to convert value in env to attribute in ContainerConfig or HostConfig of PouchContainer create api.
* pre-start container point, at this point you can set more pre-start hooks to oci spec, where you can do some special thing before container entrypoint start, priority decide the order of executing of the hook. libnetwork hook has priority 0, so if the hook is expected to run before network in container setup you should set priority to a value big then 0, and vice versa.
* pre-create-endpoint container point, at this point you can return the priority of this endpoint and if this endpoint need enable resolver and the generic params of this endpoint.
* pre-update container point, at this point you can change the input stream to your wanted format or rules.
* post-update container point, at this point you can change the update config into the container's env.

Above five points are organized by container plugins, defined as follow:

```
// ContainerPlugin defines places where a plugin will be triggered in container lifecycle
type ContainerPlugin interface {
	// PreCreate defines plugin point where receives a container create request, in this plugin point user
	// could change the container create body passed-in by http request body
	PreCreate(*types.ContainerCreateConfig) error

	// PreStart returns an array of priority and args which will pass to runc, the every priority
	// used to sort the pre start array that pass to runc, network plugin hook always has priority value 0.
	PreStart(interface{}) ([]int, [][]string, error)

	// PreCreateEndpoint accepts the container id and env of this container, to update the config of container's endpoint.
	PreCreateEndpoint(string, []string, *networktypes.Endpoint) error

	// PreUpdate defines plugin point where receives a container update request, in this plugin point user
	// could change the container update body passed-in by http request body
	PreUpdate(io.ReadCloser) (io.ReadCloser, error)

	// PostUpdate called after update method successful,
	// the method accepts the rootfs path and envs of container
	PostUpdate(string, []string) error
}
```

### daemon plugin

* pre-start daemon point, at this point you can start assistant processes like network plugins and dfget proxy which need by pouchd and whose life cycle is the same as pouchd.
* pre-stop daemon point, at this point you can stop the assistant processes gracefully, but the trigger of this point is not a promise, because pouchd may be killed by SIGKILL.

Defined as follow:

```
// DaemonPlugin defines places where a plugin will be triggered in pouchd lifecycle
type DaemonPlugin interface {
	// PreStartHook is invoked by pouch daemon before real start, in this hook user could start dfget proxy or other
	// standalone process plugins
	PreStartHook() error

	// PreStopHook is invoked by pouch daemon before daemon process exit, not a promise if daemon is killed, in this
	// hook user could stop the process or plugin started by PreStartHook
	PreStopHook() error
}
```

### volume plugin

* pre-volume-create volume point, at this point you can change the volume's create config as you want, add your default volume's options.

Defined as follow:

```
// VolumePlugin defines places where a plugin will be triggered in volume lifecycle
type VolumePlugin interface {
	// PreVolumeCreate defines plugin point where receives an volume create request, in this plugin point user
	// could change the volume create body passed-in by http request body
	PreVolumeCreate(*types.VolumeCreateConfig) error
}
```

### cri plugin

* pre-create-container cri point, at this point you can update the container config what it will be created, such as update the container's envs or labels.

Defined as follow:

```
// CriPlugin defines places where a plugin will be triggered in CRI api lifecycle
type CriPlugin interface {
	// PreCreateContainer defines plugin point where receives a container create request, in this plugin point user
	// could update the container's config in cri interface.
	PreCreateContainer(*types.ContainerCreateConfig, interface{}) error
}
```

### api plugin

* update handler, enable to add, delete, modify the PouchContainer's HTTP API handler.

Defined as follow:

```
import "github.com/alibaba/pouch/apis/server/types"
// APIPlugin provide the ability to extend PouchContainer HTTP API and change how handler behave.
type APIPlugin interface {
	// The default handler of each API would be passed in while starting HTTP server.
	// UpdateHandler could register extra HTTP API to PouchContainer server,
	// change the behavior of the default handler.
	UpdateHandler([]*types.HandlerSpec) []*types.HandlerSpec
}
```

### image plugin

* support multiple snapshotters, post pull plugin will apply content to other snapshotter

Defined as follow:

```
// ImagePlugin defines places where a plugin will be triggered in image operations
type ImagePlugin interface {
	PostPull(ctx context.Context, snapshotter string, image containerd.Image) error
}

```

## Example

### How to write

Introduce how to write a daemon plugin.

#### 1. Make your plugin package

You can make your package in directory `hookplugins/daemonplugin` and add a go file `daemon_hook.go`.

#### 2. Define a struct for your plugin

In `daemon_hook.go` define your plugin object, struct is `type daemonPlugin struct{}`.

#### 3. Register your plugin

In `init` function to register your plugin, now we provide 4 plugin to register:

* `RegisterContainerPlugin`
* `RegisterDaemonPlugin`
* `RegisterCriPlugin`
* `RegisterVolumePlugin`

In my plugin, we use `RegisterDaemonPlugin` to register a daemon plugin into pouch daemon.

```
func init() {
	hookplugins.RegisterDaemonPlugin(&daemonPlugin{})
}
```

#### 4. Implement the plugin's interface function

We implement the daemon plugin's interface function, we just print some logs.

```
// DaemonPlugin defines places where a plugin will be triggered in pouchd lifecycle
type DaemonPlugin interface {
	// PreStartHook is invoked by pouch daemon before real start, in this hook user could start http proxy or other
    // standalone process plugins
	PreStartHook() error

	// PreStopHook is invoked by pouch daemon before daemon process exit, not a promise if daemon is killed, in this
	// hook user could stop the process or plugin started by PreStartHook
	PreStopHook() error
}
```

### How to build

Use this method to build my daemon plugin:

```
# hack/module --add-plugin=github.com/alibaba/pouch/hookplugins/daemonplugin
```

And then `Makefile` will build your plugin into daemon binary, and it will be called when daemon is starting or stopping.
