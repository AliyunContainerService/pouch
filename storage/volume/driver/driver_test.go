package driver

import (
	"context"
	"testing"
)

func TestRegister(t *testing.T) {
	testDriverName := "testdriver"
	fake1Driver := NewFakeDriver(testDriverName)

	err := Register(fake1Driver)
	if err != nil {
		t.Fatal("failed to register fake1 driver")
	}

	d, err := Get(testDriverName)
	if err != nil {
		t.Fatalf("failed to get fake1 driver, err: %v", err)
	}
	if d == nil {
		t.Fatal("failed to get fake1 driver is nil")
	}

	ctx := context.Background()
	if d.Name(ctx) != testDriverName {
		t.Fatal("error driver name with testdriver")
	}

	success := Unregister(testDriverName)
	if !success {
		t.Fatal("failed to unregister testdriver")
	}

	d, err = Get(testDriverName)
	if err == nil || d != nil {
		t.Fatalf("failed to unregister testdriver, get driver: %s", d.Name(ctx))
	}
}

func TestGetAll(t *testing.T) {
	driverNames := []string{"testdriver1", "testdriver2", "testdriver3"}
	for _, name := range driverNames {
		driver := NewFakeDriver(name)
		err := Register(driver)
		if err != nil {
			t.Fatalf("failed to register driver: %s", name)
		}
	}

	defer func() {
		for _, name := range driverNames {
			success := Unregister(name)
			if !success {
				t.Fatalf("failed to unregister driver: %s", name)
			}
		}
	}()

	names := AllDriversName()
	if len(names) != 3 {
		t.Fatalf("failed to get all drivers, number is %d", len(names))
	}

	for _, n := range names {
		if n != "testdriver1" && n != "testdriver2" && n != "testdriver3" {
			t.Fatalf("failed to get all driver, name %s is unknown", n)
		}
	}
}

func TestAlias(t *testing.T) {
	driverNames := []string{"testdriver1", "testdriver2"}
	for _, name := range driverNames {
		driver := NewFakeDriver(name)
		err := Register(driver)
		if err != nil {
			t.Fatalf("failed to register driver: %s", name)
		}
	}

	defer func() {
		for _, name := range driverNames {
			success := Unregister(name)
			if !success {
				t.Fatalf("failed to unregister driver: %s", name)
			}
		}
	}()

	err := Alias("testdriver1", "testdriver111")
	if err != nil {
		t.Fatal("failed to alias driver")
	}

	d, err := Get("testdriver111")
	if err != nil {
		t.Fatal("failed to get alias volume driver.")
	}
	if d == nil {
		t.Fatal("failed to get alias volume driver")
	}

	ctx := context.Background()
	if d.Name(ctx) != "testdriver1" {
		t.Fatalf("failed to get volume name: %s", d.Name(ctx))
	}
}

func TestVolumeStoreMode_Valid(t *testing.T) {
	tests := []struct {
		Mode   VolumeStoreMode
		expect bool
	}{
		{
			LocalStore,
			true,
		},
		{
			RemoteStore,
			true,
		},
		{
			CreateDeleteInCentral,
			false,
		},
		{
			UseLocalMetaStore,
			false,
		},
		{
			LocalStore | RemoteStore,
			false,
		},
		{
			LocalStore | CreateDeleteInCentral,
			false,
		},
		{
			LocalStore | UseLocalMetaStore,
			true,
		},
		{
			RemoteStore | CreateDeleteInCentral,
			true,
		},
		{
			RemoteStore | UseLocalMetaStore,
			false,
		},
		{
			CreateDeleteInCentral | UseLocalMetaStore,
			false,
		},
		{
			LocalStore | RemoteStore | CreateDeleteInCentral,
			false,
		},
		{
			LocalStore | RemoteStore | UseLocalMetaStore,
			false,
		},
		{
			RemoteStore | CreateDeleteInCentral | UseLocalMetaStore,
			false,
		},
		{
			LocalStore | RemoteStore | CreateDeleteInCentral | UseLocalMetaStore,
			false,
		},
	}

	for index, tt := range tests {
		if tt.Mode.Valid() != tt.expect {
			t.Fatalf("failed to test valid VolumeStoreMode, index: %d, mode: %v, expect: %v",
				index+1, tt.Mode, tt.expect)
		}
	}
}
