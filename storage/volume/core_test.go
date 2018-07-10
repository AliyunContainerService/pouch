package volume

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/alibaba/pouch/storage/volume/driver"
	"github.com/alibaba/pouch/storage/volume/types"
	"github.com/stretchr/testify/assert"
)

func createVolumeCore(root string) (*Core, error) {
	cfg := Config{
		VolumeMetaPath: path.Join(root, "volume.db"),
	}

	return NewCore(cfg)
}

func createDriver(driverName string) {
	driver.Register(driver.NewFakeDriver(driverName))
}

func TestGetVolume(t *testing.T) {
	type TestCase struct {
		id types.VolumeID
	}

	testCases := []TestCase{
		{
			id: types.VolumeID{Name: "v1", Driver: "d1"},
		},
		{
			id: types.VolumeID{Name: "v2", Driver: "d2"},
		},
	}
	dir, err := ioutil.TempDir("", "TestCreateVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	core, err := createVolumeCore(dir)
	if err == nil {
		_, err := core.GetVolume(testCases[0].id)
		if err == nil {
			t.Fatal("should generate error here")
		}
	} else {
		t.Fatal("create volume core error")
	}
	driver.Register(driver.NewFakeDriver("d2"))
	defer driver.Unregister("d2")

	dir1, err := ioutil.TempDir("", "TestCreateVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1)
	core1, err := createVolumeCore(dir1)
	if err != nil {
		t.Fatal("cannot create volume core")
	}
	v, err := core1.CreateVolume(testCases[1].id)
	if err != nil {
		t.Fatal("cannot create volume")
	}

	volume, err := core1.GetVolume(testCases[1].id)
	if err != nil {
		t.Fatal("should not generate error here")
	}

	assert.Equal(t, volume.VolumeID().Name, v.VolumeID().Name)
}

func TestRemoveVolume(t *testing.T) {
	type TestCase struct {
		id types.VolumeID
	}

	testCases := []TestCase{
		{
			id: types.VolumeID{Name: "v1", Driver: "d1"},
		},
		{
			id: types.VolumeID{Name: "v2", Driver: "d2"},
		},
	}
	dir, err := ioutil.TempDir("", "TestRemoveVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	core, err := createVolumeCore(dir)
	if err == nil {
		err := core.RemoveVolume(testCases[0].id)
		if err == nil {
			t.Fatal("should generate error here, cannot remove unexisted volume")
		}
	} else {
		t.Fatal("create volume core error")
	}

	driver.Register(driver.NewFakeDriver("d2"))
	defer driver.Unregister("d2")

	dir1, err := ioutil.TempDir("", "TestRemoveVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1)
	core1, coreerr := createVolumeCore(dir1)
	if coreerr != nil {
		t.Fatal("cannot create volume core")
	}
	_, volumeerr := core1.CreateVolume(testCases[1].id)
	if volumeerr != nil {
		t.Fatal("cannot create volume")
	}

	removeerror := core1.RemoveVolume(testCases[1].id)
	if removeerror != nil {
		t.Fatal("should not generate error here")
	}
}

func TestVolumePath(t *testing.T) {
	type TestCase struct {
		id types.VolumeID
	}

	testCases := []TestCase{
		{
			id: types.VolumeID{Name: "v1", Driver: "d1"},
		},
		{
			id: types.VolumeID{Name: "v2", Driver: "d2"},
		},
	}
	dir, err := ioutil.TempDir("", "TestVolumePath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	core, err := createVolumeCore(dir)
	if err == nil {
		_, err := core.VolumePath(testCases[0].id)
		if err == nil {
			t.Fatal("should generate error here, cannot get path of unexisted volume")
		}
	} else {
		t.Fatal("create volume core error")
	}
	driver2 := driver.NewFakeDriver("d2")
	driver.Register(driver2)
	defer driver.Unregister("d2")

	dir1, err := ioutil.TempDir("", "TestVolumePath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1)
	core1, coreerr := createVolumeCore(dir1)
	if coreerr != nil {
		t.Fatal("cannot create volume core")
	}
	volume2, volumeerr := core1.CreateVolume(testCases[1].id)
	if volumeerr != nil {
		t.Fatal("cannot create volume")
	}

	path, patherror := core1.VolumePath(testCases[1].id)
	if patherror != nil {
		t.Fatal("should not generate error here, should get path of volume")
	}
	driverPath, driverErr := driver2.Path(driver.Contexts(), volume2)
	if driverErr != nil {
		t.Fatal("cannot get volume's path")
	}
	assert.Equal(t, driverPath, path)
}

func TestAttachVolume(t *testing.T) {
	type TestCase struct {
		id types.VolumeID
	}

	testCases := []TestCase{
		{
			id: types.VolumeID{Name: "v1", Driver: "d1"},
		},
		{
			id: types.VolumeID{Name: "v2", Driver: "d2"},
		},
	}
	dir, err := ioutil.TempDir("", "TestVolumePath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	core, err := createVolumeCore(dir)
	if err == nil {
		_, err := core.VolumePath(testCases[0].id)
		if err == nil {
			t.Fatal("should generate error here, cannot get path of unexisted volume")
		}
	} else {
		t.Fatal("create volume core error")
	}
	driver2 := driver.NewFakeDriver("d2")
	driver.Register(driver2)
	defer driver.Unregister("d2")

	dir1, err := ioutil.TempDir("", "TestVolumePath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1)
	core1, coreerr := createVolumeCore(dir1)
	if coreerr != nil {
		t.Fatal("cannot create volume core")
	}
	volume2, volumeerr := core1.CreateVolume(testCases[1].id)
	if volumeerr != nil {
		t.Fatal("cannot create volume")
	}

	path, patherror := core1.VolumePath(testCases[1].id)
	if patherror != nil {
		t.Fatal("should not generate error here, should get path of volume")
	}
	driverPath, driverErr := driver2.Path(driver.Contexts(), volume2)
	if driverErr != nil {
		t.Fatal("cannot get volume's path")
	}
	assert.Equal(t, driverPath, path)
}
