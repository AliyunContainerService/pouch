# Introduction for volume

## What is Volume

Volume is a universal module that used to provide container storage in PouchContainer. It provides storage service for containers through the interface of file storage.

## What is the architecture of Volume

Volume includes these modules is as following:

* VolumeManager: volume manager provides the basic volume interface functions for pouchd.
* Core: volume's core module, it is used to associate with several modules, and it achieves a common process that volume operate functions.
* Driver: it is used to abstract the basic functions of volume driver.
* Modules: Different types of storage provides different modules, it achieves different storage unified access to the PouchContainer volume.

The relationship between each module is as following:

![pouch_volume_architecture | center | 710x515 ](../docs/static_files/pouch_volume_architecture.png)

### VolumeManager

It provides interface is as following:

```go
type VolumeMgr interface {
    // Create is used to create volume.
    Create(ctx context.Context, name, driver string, options, labels map[string]string) (*types.Volume, error)

    // Remove is used to remove an existing volume.
    Remove(ctx context.Context, name string) error

    // List returns all volumes on this host.
    List(ctx context.Context, labels map[string]string) ([]string, error)

    // Get returns the information of volume that specified name/id.
    Get(ctx context.Context, name string) (*types.Volume, error)

    // Path returns the mount path of volume.
    Path(ctx context.Context, name string) (string, error)

    // Attach is used to bind a volume to container.
    Attach(ctx context.Context, name string, options map[string]string) (*types.Volume, error)

    // Detach is used to unbind a volume from container.
    Detach(ctx context.Context, name string, options map[string]string) (*types.Volume, error)
}
```

### Core

Core provides these functions is as following:

```go
// GetVolume return a volume's info with specified name, If not errors.
func (c *Core) GetVolume(id types.VolumeID) (*types.Volume, error)

// CreateVolume use to create a volume, if failed, will return error info.
func (c *Core) CreateVolume(id types.VolumeID) error

// ListVolumeName return the name of all volumes only.
// Param 'labels' use to filter the volume's names, only return those you want.
func (c *Core) ListVolumeName(labels map[string]string) ([]string, error)

// RemoveVolume remove volume from storage and meta information, if not success return error.
func (c *Core) RemoveVolume(id types.VolumeID) error

// VolumePath return the path of volume on node host.
func (c *Core) VolumePath(id types.VolumeID) (string, error)

// AttachVolume to enable a volume on local host.
func (c *Core) AttachVolume(id types.VolumeID, extra map[string]string) (*types.Volume, error)

// DetachVolume to disable a volume on local host.
func (c *Core) DetachVolume(id types.VolumeID, extra map[string]string) (*types.Volume, error)
```

### Driver

Driver layer provides two types of interfaces, one is the basic interfaces that all modules must implement, the others are optional interfaces for accessing differences provided by different types of storage.

* Basic interfaces

```go
type Driver interface {
    // Name returns backend driver's name.
    Name(Context) string

    // StoreMode defines backend driver's store model.
    StoreMode(Context) VolumeStoreMode

    // Create a volume.
    Create(Context, *types.Volume) error

    // Remove a volume.
    Remove(Context, *types.Volume) error

    // Path returns volume's path.
    Path(Context, *types.Volume) (string, error)
}
```

* Optional interfaces

```go
// Opt represents volume driver option interface.
type Opt interface {
    // Options return module customize volume options.
    Options() map[string]types.Option
}

// AttachDetach represents volume attach/detach interface.
type AttachDetach interface {
    // Attach a Volume to host, enable the volume.
    Attach(Context, *types.Volume) error

    // Detach a volume with host, disable the volume.
    Detach(Context, *types.Volume) error
}

// Formator represents volume format interface.
type Formator interface {
    // Format a volume.
    Format(Context, *types.Volume) error
}

```

### Modules

As of now, PouchContainer volume supports the following types of storage: local, tmpfs, ceph. If you want to add a new driver, you can refer to the sample code: [demo](volume/examples/demo/demo.go)

## How to use volume

As of now, volume supports the following operations: create/remove/list/inspect, for more details, please refer: [Volume Cli](docs/commandline/pouch_volume.md)

## Volume roadmap

PouchContainer volume will implement the interface of [CSI(container storage interface)](https://github.com/container-storage-interface/spec), as a node-server integrating with control-server to achieve volume scheduling ability.
