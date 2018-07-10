package volume

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/alibaba/pouch/storage/volume/driver"
	volerr "github.com/alibaba/pouch/storage/volume/error"
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
	// TODO
}

func TestListVolumes(t *testing.T) {
	// TODO
}

func TestListVolumeName(t *testing.T) {
	// TODO
}

func TestRemoveVolume(t *testing.T) {
	// TODO
}

func TestVolumePath(t *testing.T) {
	// TODO
}

func TestAttachVolume(t *testing.T) {
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

	driverName1 := "fake1"
	volumeName1 := "test1"
	vID1 := types.VolumeID{Name: volumeName1, Driver: driverName1}
	driver.Register(driver.NewFakeDriver(volumeDriverName))
	defer driver.Unregister(volumeDriverName)

	extra := map[string]string{}

	v0, err0 := core.AttachVolume(vID1, extra)
	if v0 != nil {
		t.Fatalf("expect get volume nil, but got a volume with name %s", v0.Name)
	}
	if err0 != volerr.ErrVolumeNotFound {
		if err0 == nil {
			t.Fatal("expect get volume not found error, but err is nil")
		} else {
			t.Fatalf("expect get volume not found error, but got %v", err0)
		}
	}

	core.CreateVolume(types.VolumeID{Name: "test1", Driver: volumeDriverName})

	v1, err1 := core.AttachVolume(vID1, extra)
	if err1 != nil {
		t.Fatalf("attach volume error: %v", err1)
	}

	if v1.Name != volumeName1 {
		t.Fatalf("expect volume name is %s, but got %s", volumeName1, v1.Name)
	}
	if v1.Driver() != driverName1 {
		t.Fatalf("expect volume driver is %s, but got %s", driverName1, v1.Driver())
	}
}

func TestDetachVolume(t *testing.T) {
	// TODO
}
