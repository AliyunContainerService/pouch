package volume

import (
	"io/ioutil"
	"os"
	"path"
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
	// TODO
}

func TestListVolumes(t *testing.T) {
	// TODO
}

func TestListVolumeName(t *testing.T) {
	// TODO
}

func TestRemoveVolume(t *testing.T) {
	volName1 := "vol2"
	driverName1 := "fake_driver12"
	volid1 := types.VolumeID{Name: volName1, Driver: driverName1}

	dir, err := ioutil.TempDir("", "TestGetVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	core, err := createVolumeCore(dir)
	if err != nil {
		t.Fatal(err)
	}

	driver.Register(driver.NewFakeDriver(driverName1))
	defer driver.Unregister(driverName1)

	v1, err1 := core.CreateVolume(volid1)
	if err != nil {
		t.Fatalf("create volume error: %v", err1)
	}
	if v1.Name != volName1 {
		t.Fatalf("expect volume name is %s, but got %s", volName1, v1.Name)
	}
	if v1.Driver() != driverName1 {
		t.Fatalf("expect volume driver is %s, but got %s", driverName1, v1.Driver())
	}

	err2 := core.RemoveVolume(volid1)
	if err2 != nil {
		t.Fatalf("remove volume id %v error: %v", volid1, err2)
	}

	err3 := core.RemoveVolume(volid1)
	if err3 == nil {
		t.Fatalf("expect remove empty volume id %v error, but return nil", volid1)
	}
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
