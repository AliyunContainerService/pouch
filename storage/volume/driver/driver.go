package driver

import (
	"fmt"
	"regexp"
	"sort"
	"sync"

	"github.com/alibaba/pouch/plugins"
	"github.com/pkg/errors"
)

const (
	// Driver's name can only contain these characters, a-z A-Z 0-9.
	driverNameRegexp = "^[a-zA-Z0-9].*$"
	// Option's name can only contain these characters, a-z 0-9 - _.
	optionNameRegexp = "^[a-z0-9-_].*$"
	// volumePluginType is the plugin which implements volume driver
	volumePluginType = "VolumeDriver"
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

// driverTable contains all volume drivers
type driverTable struct {
	sync.Mutex
	drivers map[string]Driver
}

// Get is used to get driver from driver table.
func (t *driverTable) Get(name string) (Driver, error) {
	t.Lock()
	v, ok := t.drivers[name]
	if ok {
		t.Unlock()
		return v, nil
	}
	t.Unlock()

	plugin, err := plugins.Get(volumePluginType, name)
	if err != nil {
		return nil, fmt.Errorf("%s driver not found: %v", name, err)
	}

	driver := NewRemoteDriverWrapper(name, plugin)

	t.Lock()
	defer t.Unlock()

	v, ok = t.drivers[name]
	if !ok {
		v = driver
		t.drivers[name] = v
	}

	return v, nil
}

// GetAll will list all volume drivers.
func (t *driverTable) GetAll() ([]Driver, error) {
	pluginList, err := plugins.GetAll(volumePluginType)
	if err != nil {
		return nil, fmt.Errorf("error listing plugins: %v", err)
	}

	var driverList []Driver

	t.Lock()
	defer t.Unlock()

	for _, d := range t.drivers {
		driverList = append(driverList, d)
	}

	for _, p := range pluginList {
		_, ok := t.drivers[p.Name]
		if ok {
			// the driver has existed, ignore it.
			continue
		}

		d := NewRemoteDriverWrapper(p.Name, p)

		t.drivers[p.Name] = d
		driverList = append(driverList, d)
	}

	return driverList, nil
}

var backendDrivers = &driverTable{
	drivers: make(map[string]Driver),
}

// Register add a backend driver module.
func Register(d Driver) error {
	ctx := Contexts()
	driverName := d.Name(ctx)

	matched, err := regexp.MatchString(driverNameRegexp, driverName)
	if err != nil {
		return err
	}
	if !matched {
		return errors.Errorf("Invalid driver name: %s, not match: %s", driverName, driverNameRegexp)
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

	backendDrivers.Lock()
	defer backendDrivers.Unlock()

	if _, ok := backendDrivers.drivers[driverName]; ok {
		return errors.Errorf("backend driver's name \"%s\" duplicate", driverName)
	}

	backendDrivers.drivers[driverName] = d
	return nil
}

// Unregister deletes a driver from driverTable
func Unregister(name string) bool {
	backendDrivers.Lock()
	defer backendDrivers.Unlock()

	_, exist := backendDrivers.drivers[name]
	if !exist {
		return false
	}

	delete(backendDrivers.drivers, name)

	return true
}

// Get returns one backend driver with specified name.
func Get(name string) (Driver, error) {
	return backendDrivers.Get(name)
}

// GetAll returns all volume drivers.
func GetAll() ([]Driver, error) {
	return backendDrivers.GetAll()
}

// Exist return true if the backend driver is registered.
func Exist(name string) bool {
	_, err := Get(name)
	if err != nil {
		return false
	}

	return true
}

// AllDriversName return all registered backend driver's name.
func AllDriversName() []string {
	// probing all volume plugins.
	backendDrivers.GetAll()

	backendDrivers.Lock()
	defer backendDrivers.Unlock()

	var names []string
	for n := range backendDrivers.drivers {
		names = append(names, n)
	}

	// sort the names.
	sort.Strings(names)

	return names
}

// Alias is used to add driver name's alias into exist driver.
func Alias(name, alias string) error {
	matched, err := regexp.MatchString(driverNameRegexp, alias)
	if err != nil {
		return err
	}
	if !matched {
		return errors.Errorf("Invalid driver name: %s, not match: %s", name, driverNameRegexp)
	}

	backendDrivers.Lock()
	defer backendDrivers.Unlock()

	// check whether the driver exists
	d, ok := backendDrivers.drivers[name]
	if !ok {
		return errors.Errorf("volume driver: %s is not exist", name)
	}

	// alias should not exist
	_, ok = backendDrivers.drivers[alias]
	if ok {
		return errors.Errorf("Invalid volume alias: %s, duplicate", name)
	}

	backendDrivers.drivers[alias] = d

	return nil
}
