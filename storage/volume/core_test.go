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
  
	volumeDriverName := "fake1"
	driver.Register(driver.NewFakeDriver(volumeDriverName))
	defer driver.Unregister(volumeDriverName)
  
	// Test1: get volume return nil
	_, err = core.GetVolume(types.VolumeID{Name: "testGetVolume", Driver: volumeDriverName})
	if err == nil {
	  t.Fatal("expect get driver not found error, but err is nil")
	}
  }
  
  func TestListVolumes(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestListVolumes")
	if err != nil {
	  t.Fatal(err)
	}
	defer os.RemoveAll(dir)
  
	// create volume core
	core, err := createVolumeCore(dir)
	if err != nil {
	  t.Fatal(err)
	}
  
	volumeDriverName := "fake1"
	driver.Register(driver.NewFakeDriver(volumeDriverName))
	defer driver.Unregister(volumeDriverName)
  
	// Test1
	_, err = core.ListVolumes(nil)
	if err != nil {
	  t.Fatal("expect get driver not found error, but err is nil")
	}
  }
  
  func TestListVolumeName(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestListVolumeName")
	if err != nil {
	  t.Fatal(err)
	}
	defer os.RemoveAll(dir)
  
	// create volume core
	core, err := createVolumeCore(dir)
	if err != nil {
	  t.Fatal(err)
	}
  
	volumeDriverName := "fake1"
	driver.Register(driver.NewFakeDriver(volumeDriverName))
	defer driver.Unregister(volumeDriverName)
  
	// Test1
	_, err = core.ListVolumeName(nil)
	if err != nil {
	  t.Fatal("expect get driver not found error, but err is nil")
	}
  }
  
  func TestRemoveVolume(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestRemoveVolume")
	if err != nil {
	  t.Fatal(err)
	}
	defer os.RemoveAll(dir)
  
	// create volume core
	core, err := createVolumeCore(dir)
	if err != nil {
	  t.Fatal(err)
	}
  
	volumeDriverName := "fake1"
	driver.Register(driver.NewFakeDriver(volumeDriverName))
	defer driver.Unregister(volumeDriverName)
  
	// Test1
	err = core.RemoveVolume(types.VolumeID{Name: "test1", Driver: volumeDriverName})
	if err == nil {
	  t.Fatal("expect get driver not found error, but err is nil")
	}
  }
  
  func TestVolumePath(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestVolumePath")
	if err != nil {
	  t.Fatal(err)
	}
	defer os.RemoveAll(dir)
  
	// create volume core
	core, err := createVolumeCore(dir)
	if err != nil {
	  t.Fatal(err)
	}
  
	volumeDriverName := "fake1"
	driver.Register(driver.NewFakeDriver(volumeDriverName))
	defer driver.Unregister(volumeDriverName)
  
	// Test1: return nil
	_, err = core.VolumePath(types.VolumeID{Name: "test1", Driver: volumeDriverName})
	if err == nil {
	  t.Fatalf("VolumePath error: %v", err)
	}
  }
  
  func TestAttachVolume(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestAttachVolume")
	if err != nil {
	  t.Fatal(err)
	}
	defer os.RemoveAll(dir)
  
	// create volume core
	core, err := createVolumeCore(dir)
	if err != nil {
	  t.Fatal(err)
	}
  
	volumeDriverName := "fake1"
	driver.Register(driver.NewFakeDriver(volumeDriverName))
	defer driver.Unregister(volumeDriverName)
  
	vleID := types.VolumeID{Name: "test1", Driver: volumeDriverName}
	// Test1: return nil
	m := map[string]string{
	  "k1": "v1",
	  "k2": "v2",
	}
	_, err = core.AttachVolume(vleID, m)
	if err == nil {
	  t.Fatalf("AttachVolume error: %v", err)
	}
  
	// Test2 
	core.CreateVolume(vleID)
	v, err := core.AttachVolume(vleID, m)
	if err != nil {
	  t.Fatalf("AttachVolume error: %v", err)
	}
  }
  
  func TestDetachVolume(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestDetachVolume")
	if err != nil {
	  t.Fatal(err)
	}
	defer os.RemoveAll(dir)
  
	// create volume core
	core, err := createVolumeCore(dir)
	if err != nil {
	  t.Fatal(err)
	}
  
	volumeDriverName := "fake1"
	driver.Register(driver.NewFakeDriver(volumeDriverName))
	defer driver.Unregister(volumeDriverName)
  
	// Test1: return nil
	m := map[string]string{
	  "k1": "v1",
	  "k2": "v2",
	}
	_, err = core.DetachVolume(types.VolumeID{Name: "test1", Driver: volumeDriverName}, m)
	if err == nil {
	  t.Fatalf("DetachVolume error: %v", err)
	}
  }