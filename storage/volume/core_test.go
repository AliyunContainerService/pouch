package volume

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
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

	dir, err := ioutil.TempDir("", "TestGetVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// create volume core
	core, err := createVolumeCore(dir)
	if err != nil {
		t.Fatal(err)
	}

	// add one volume and get
	driverName1 := "fake1"
	volumeName1 := "test1"
	vID1 := types.VolumeID{Name: volumeName1, Driver: driverName1}
	driver.Register(driver.NewFakeDriver(driverName1))
	defer driver.Unregister(driverName1)

	v0, err0 := core.GetVolume(vID1)
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

	core.CreateVolume(vID1)

	v1, err1 := core.GetVolume(vID1)
	if err1 != nil {
		t.Fatalf("get volume error: %v", err1)
	}

	if v1.Name != volumeName1 {
		t.Fatalf("expect volume name is %s, but got %s", volumeName1, v1.Name)
	}
	if v1.Driver() != driverName1 {
		t.Fatalf("expect volume driver is %s, but got %s", driverName1, v1.Driver())
	}

	// add two volumes and get
	driverName2 := "fake1"
	volumeName2 := "test1"
	vID2 := types.VolumeID{Name: volumeName2, Driver: driverName2}
	driver.Register(driver.NewFakeDriver(driverName2))
	defer driver.Unregister(driverName2)

	core.CreateVolume(vID2)

	v2, err2 := core.GetVolume(vID2)
	if err2 != nil {
		t.Fatalf("get volume error: %v", err2)
	}

	if v2.Name != volumeName2 {
		t.Fatalf("expect volume name is %s, but got %s", volumeName2, v2.Name)
	}
	if v2.Driver() != driverName2 {
		t.Fatalf("expect volume driver is %s, but got %s", driverName2, v2.Driver())
	}

	v2_1, err2_1 := core.GetVolume(vID1)
	if err2_1 != nil {
		t.Fatalf("get volume error: %v", err2_1)
	}

	if v2_1.Name != volumeName1 {
		t.Fatalf("expect volume name is %s, but got %s", volumeName1, v2_1.Name)
	}
	if v2_1.Driver() != driverName1 {
		t.Fatalf("expect volume driver is %s, but got %s", driverName1, v2_1.Driver())
	}
}

func TestListVolumes(t *testing.T) {
	driverName := "fake_driver4"
	dir, err := ioutil.TempDir("", "TestGetVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	core, err := createVolumeCore(dir)
	if err != nil {
		t.Fatal(err)
	}

	driver.Register(driver.NewFakeDriver(driverName))
	defer driver.Unregister(driverName)

	var i int64
	volmap := map[string]*types.Volume{}
	for i = 0; i < 6; i++ {
		volName := strconv.FormatInt(i, 10)
		volid := types.VolumeID{Name: volName, Driver: driverName}
		v, err := core.CreateVolume(volid)
		if err != nil {
			t.Fatalf("create volume error: %v", err)
		}
		volmap[volName] = v
	}

	volarray, err := core.ListVolumes(nil)
	for k := 0; k < len(volarray); k++ {
		vol := volarray[k]
		_, found := volmap[vol.Name]
		if found == false {
			t.Fatalf("list volumes %v not found", vol)
		}
	}
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
	volName1 := "vol3"
	driverName1 := "fake_driver"
	volid1 := types.VolumeID{Name: volName1, Driver: driverName1}
	
	expectPath := path.Join("/fake/", volName1) //keep consist with the path API in fake_driver.go
	dir, err := ioutil.TempDir("", "TestVolumePath")
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

	path1, err := core.VolumePath(volid1)
	if err == nil {
		t.Fatalf("expect volume not found err when get volume path from empty volume, but get path: %s", path1)
	}

	_, err1 := core.CreateVolume(volid1)
	if err1 != nil {
		t.Fatal(err1)
	}

	path2, err := core.VolumePath(volid1)
	if  err != nil {
		t.Fatalf("get volume path error: %v", err)
	}

	if path2 != expectPath {
		t.Fatalf("expect volume path is %s, but got %s", expectPath, path2)
	}
}

func TestAttachVolume(t *testing.T) {
	// TODO
}

func TestDetachVolume(t *testing.T) {
	// TODO
}
