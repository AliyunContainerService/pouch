package driver

import (
	"regexp"

	"github.com/alibaba/pouch/storage/volume/types"

	"github.com/pkg/errors"
)

const (
	// Driver's name can only contain these characters, a-z A-Z 0-9.
	driverNameRegexp = "^[a-zA-Z0-9].*$"
	// Option's name can only contain these characters, a-z 0-9 - _.
	optionNameRegexp = "^[a-z0-9-_].*$"
)

const (
	// LocalStore defines Local store mode.
	LocalStore VolumeStoreMode = 1

	// RemoteStore defines remote store mode.
	RemoteStore VolumeStoreMode = 2

	// CreateDeleteInCentral defines operate storage on gateway.
	CreateDeleteInCentral VolumeStoreMode = 4

	// UseLocalMetaStore defines store metadata on local host.
	UseLocalMetaStore VolumeStoreMode = 8
)

// VolumeStoreMode defines volume store mode type.
type VolumeStoreMode int

// String returns VolumeStoreMode with string description.
func (m VolumeStoreMode) String() string {
	return ""
}

// Valid is used to check VolumeStoreMode is valid or not.
func (m VolumeStoreMode) Valid() bool {
	if m.IsLocal() && m.IsRemote() {
		return false
	}

	// local store
	if m.IsLocal() {
		if m.CentralCreateDelete() {
			return false
		}
		return true
	}

	// remote store
	if m.IsRemote() {
		if m.UseLocalMeta() {
			return false
		}
		return true
	}

	return false
}

// IsRemote checks VolumeStoreMode is remote mode or not.
func (m VolumeStoreMode) IsRemote() bool {
	return (m & RemoteStore) != 0
}

// IsLocal checks VolumeStoreMode is local mode or not.
func (m VolumeStoreMode) IsLocal() bool {
	return (m & LocalStore) != 0
}

// CentralCreateDelete checks VolumeStoreMode is center mode of operation or not.
func (m VolumeStoreMode) CentralCreateDelete() bool {
	return (m & CreateDeleteInCentral) != 0
}

// UseLocalMeta checks VolumeStoreMode is store metadata on local or not.
func (m VolumeStoreMode) UseLocalMeta() bool {
	return (m & UseLocalMetaStore) != 0
}

//
type driverTable map[string]Driver

// Add is used to add driver into driver table.
func (t driverTable) Add(name string, d Driver) {
	t[name] = d
}

// Del is used to delete driver from driver table.
func (t driverTable) Del(name string) {
	delete(t, name)
}

// Get is used to get driver from driver table.
func (t driverTable) Get(name string) (Driver, bool) {
	v, ok := t[name]
	return v, ok
}

var backendDrivers driverTable

// Register add a backend driver module.
func Register(d Driver) error {
	if backendDrivers == nil {
		backendDrivers = make(driverTable)
	}
	ctx := Contexts()

	matched, err := regexp.MatchString(driverNameRegexp, d.Name(ctx))
	if err != nil {
		return err
	}
	if !matched {
		return errors.Errorf("Invalid driver name: %s, not match: %s", d.Name(ctx), driverNameRegexp)
	}

	if _, ok := backendDrivers.Get(d.Name(ctx)); ok {
		return errors.Errorf("Backend driver's name \"%s\" duplicate", d.Name(ctx))
	}

	if !d.StoreMode(ctx).Valid() {
		return errors.Errorf("Invalid driver store mode: %d", d.StoreMode(ctx))
	}

	if opt, ok := d.(Opt); ok {
		for name, opt := range opt.Options() {
			matched, err := regexp.MatchString(optionNameRegexp, name)
			if err != nil {
				return err
			}
			if !matched || opt.Desc == "" {
				return errors.Errorf("Invalid option name: %s or desc: %s", name, opt.Desc)
			}
		}
	}

	backendDrivers.Add(d.Name(ctx), d)
	return nil
}

// Get return one backend driver with specified name.
func Get(name string) (Driver, bool) {
	return backendDrivers.Get(name)
}

// Exist return true if the backend driver is registered.
func Exist(name string) bool {
	_, ok := Get(name)
	return ok
}

// List return all backend drivers.
func List() []Driver {
	drivers := make([]Driver, 0, 5)
	for _, d := range backendDrivers {
		drivers = append(drivers, d)
	}
	return drivers
}

// AllDriversName return all backend driver's name.
func AllDriversName() []string {
	names := make([]string, 0, 5)
	for n := range backendDrivers {
		names = append(names, n)
	}
	return names
}

// ListDriverOption return backend driver's options by name.
func ListDriverOption(name string) map[string]types.Option {
	dv, ok := Get(name)
	if !ok {
		return nil
	}
	if opt, ok := dv.(Opt); ok {
		return opt.Options()
	}
	return nil
}

// Alias is used to add driver name's alias into exist driver.
func Alias(name, alias string) error {
	d, exist := backendDrivers.Get(name)
	if !exist {
		return errors.Errorf("volume driver: %s is not exist", name)
	}

	matched, err := regexp.MatchString(driverNameRegexp, alias)
	if err != nil {
		return err
	}
	if !matched {
		return errors.Errorf("Invalid driver name: %s, not match: %s", name, driverNameRegexp)
	}

	backendDrivers.Add(alias, d)

	return nil
}
