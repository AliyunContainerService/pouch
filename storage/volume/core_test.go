package volume

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"

	"github.com/alibaba/pouch/storage/volume/driver"
	"github.com/alibaba/pouch/storage/volume/types"
)

func createVolumeCore(root string) (*Core, error) {
	cfg := Config{
		VolumeMetaPath: path.Join(root, "volume.db"),
	}

	return NewCore(cfg)
}

func TestCreateVolume(t *testing.T) {
	volumeDriverName := "fake1"

	dir, err := ioutil.TempDir("", "TestCreateVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// create volume core
	core, err := createVolumeCore(dir)
	if err != nil {
		t.Fatal(err)
	}

	driver.Register(driver.NewFakeDriver(volumeDriverName))
	defer driver.Unregister(volumeDriverName)

	v, err := core.CreateVolume(types.VolumeID{Name: "test1", Driver: volumeDriverName})
	if err != nil {
		t.Fatalf("create volume error: %v", err)
	}

	if v.Name != "test1" {
		t.Fatalf("expect volume name is %s, but got %s", "test1", v.Name)
	}
	if v.Driver() != volumeDriverName {
		t.Fatalf("expect volume driver is %s, but got %s", volumeDriverName, v.Driver())
	}

	_, err = core.CreateVolume(types.VolumeID{Name: "none", Driver: "none"})
	if err == nil {
		t.Fatal("expect get driver not found error, but err is nil")
	}
}
func TestGetVolume(t *testing.T) {
	volumeDriverName := "fake1"
	dir, err := ioutil.TempDir("", "TestGetVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	core, err := createVolumeCore(dir)
	if err != nil {
		t.Fatal(err)
	}

	params := types.VolumeID{Name: "test1", Driver: volumeDriverName}
	// before backend module registered
	v, e := core.GetVolume(params)
	if e == nil {
		t.Fatal("expect metastore ErrObjectNotFound err, but get nil err")
	}

	driver.Register(driver.NewFakeDriver(volumeDriverName))
	defer driver.Unregister(volumeDriverName)

	_, err = core.CreateVolume(params)
	if err != nil {
		t.Fatalf("create volume error: %v", err)
	}

	// after volume created
	v, e = core.GetVolume(params)
	if e != nil {
		t.Fatal("expect nil err, but get err %s", e)
	}
	if v == nil {
		t.Fatal("expect volume not nil, but nil volume")
	}

	if v.Name != "test1" {
		t.Fatalf("expect volume name is %s, but got %s", "test1", v.Name)
	}

	if v.Driver() != volumeDriverName {
		t.Fatalf("expect volume driver is %s, but got %s", volumeDriverName, v.Driver())
	}
}

func TestListVolumeName(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestListVolumeName")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	core, err := createVolumeCore(dir)
	if err != nil {
		t.Fatal(err)
	}

	ids := []types.VolumeID{
		{Name: "test1", Driver: "fake1"},
		{Name: "test2", Driver: "fake2"},
	}

	l := []string{
		ids[0].Name, ids[1].Name,
	}

	ns, e := core.ListVolumeName(nil)
	if e != nil {
		t.Fatalf("expect nil err, but get err %s", e)
	}
	if ns != nil {
		t.Fatal("expect name list nil, but get %s", ns)
	}

	// register backend module
	driver.Register(driver.NewFakeDriver(ids[0].Driver))
	defer driver.Unregister(ids[0].Driver)

	_, err = core.CreateVolume(ids[0])
	if err != nil {
		t.Fatalf("create volume error: %s", err)
	}

	ns, e = core.ListVolumeName(nil)
	if e != nil {
		t.Fatalf("expect nil err, but get err %s", e)
	}
	if ns == nil {
		t.Fatal("expect name list not nil, but get nil")
	}
	if len(ns) != 1 {
		t.Fatalf("expect name list of length 1, but get %v", len(ns))
	}
	if ns[0] != ids[0].Name {
		t.Fatalf("expect name %s, but get %s", ids[0].Name, ns[0])
	}

	// register backend module
	driver.Register(driver.NewFakeDriver(ids[1].Driver))
	defer driver.Unregister(ids[1].Driver)

	_, err = core.CreateVolume(ids[1])
	if err != nil {
		t.Fatalf("create volume error: %s", err)
	}

	ns, e = core.ListVolumeName(nil)
	if e != nil {
		t.Fatalf("expect nil err, but get err %s", e)
	}
	if ns == nil {
		t.Fatal("expect name list not nil, but get nil")
	}
	if len(ns) != 2 {
		t.Fatalf("expect name list of length 12, but get %v", len(ns))
	}
	sort.Strings(ns)
	sort.Strings(l)
	if !reflect.DeepEqual(ns, l) {
		t.Errorf("ListVolumeName() = %s, want %s", ns, l)
	}
}

func TestVolumeLists(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestVolumeLists")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	core, err := createVolumeCore(dir)
	if err != nil {
		t.Fatal(err)
	}

	ids := []types.VolumeID{
		{Name: "test1", Driver: "fake1"},
		{Name: "test2", Driver: "fake2"},
	}

	// before backend module registered
	v, e := core.ListVolumes(nil)
	if e != nil {
		t.Fatalf("expect nil err, but get err %s", e)
	}
	if v == nil {
		t.Fatal("expect volumes not nil, but get nil")
	}
	if len(v) != 0 {
		t.Fatalf("expect volumes of length 0, but get %v", len(v))
	}

	// register one module
	driver.Register(driver.NewFakeDriver(ids[0].Driver))
	defer driver.Unregister(ids[0].Driver)

	_, err = core.CreateVolume(ids[0])
	if err != nil {
		t.Fatalf("create volume error: %s", err)
	}

	v, e = core.ListVolumes(nil)
	if e != nil {
		t.Fatalf("expect nil err, but get err %s", e)
	}
	if v == nil {
		t.Fatal("expect volumes not nil, but get nil")
	}
	if len(v) != 1 {
		t.Fatalf("expect volumes of length 1, but get %v", len(v))
	}

	// register another module
	driver.Register(driver.NewFakeDriver(ids[1].Driver))
	defer driver.Unregister(ids[1].Driver)

	_, err = core.CreateVolume(ids[1])
	if err != nil {
		t.Fatalf("create volume error: %s", err)
	}

	v, e = core.ListVolumes(nil)
	if e != nil {
		t.Fatalf("expect nil err, but get err %s", e)
	}
	if v == nil {
		t.Fatal("expect volumes not nil, but get nil")
	}
	if len(v) != 2 {
		t.Fatalf("expect volumes of length 2, but get %v", len(v))
	}

}
func TestRemoveVolume(t *testing.T) {
	// TODO
}

func TestVolumePath(t *testing.T) {
	// TODO
}

func TestAttachVolume(t *testing.T) {
	// TODO
}

func TestDetachVolume(t *testing.T) {
	// TODO
}
