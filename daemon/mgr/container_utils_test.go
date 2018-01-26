package mgr

import (
	"reflect"
	"testing"

	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/pkg/collect"
	"github.com/docker/libnetwork/config"
)

func TestContainerManager_containerID(t *testing.T) {
	type fields struct {
		Store         *meta.Store
		Client        *ctrd.Client
		NameToID      *collect.SafeMap
		ImageMgr      ImageMgr
		VolumeMgr     VolumeMgr
		NetworkMgr    NetworkMgr
		IOs           *containerio.Cache
		ExecProcesses *collect.SafeMap
		Config        *config.Config
		cache         *collect.SafeMap
		monitor       *ContainerMonitor
	}
	type args struct {
		nameOrPrefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &ContainerManager{
				Store:         tt.fields.Store,
				Client:        tt.fields.Client,
				NameToID:      tt.fields.NameToID,
				ImageMgr:      tt.fields.ImageMgr,
				VolumeMgr:     tt.fields.VolumeMgr,
				NetworkMgr:    tt.fields.NetworkMgr,
				IOs:           tt.fields.IOs,
				ExecProcesses: tt.fields.ExecProcesses,
				Config:        tt.fields.Config,
				cache:         tt.fields.cache,
				monitor:       tt.fields.monitor,
			}
			got, err := mgr.containerID(tt.args.nameOrPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerManager.containerID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ContainerManager.containerID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainerManager_container(t *testing.T) {
	type fields struct {
		Store         *meta.Store
		Client        *ctrd.Client
		NameToID      *collect.SafeMap
		ImageMgr      ImageMgr
		VolumeMgr     VolumeMgr
		NetworkMgr    NetworkMgr
		IOs           *containerio.Cache
		ExecProcesses *collect.SafeMap
		Config        *config.Config
		cache         *collect.SafeMap
		monitor       *ContainerMonitor
	}
	type args struct {
		nameOrPrefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Container
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &ContainerManager{
				Store:         tt.fields.Store,
				Client:        tt.fields.Client,
				NameToID:      tt.fields.NameToID,
				ImageMgr:      tt.fields.ImageMgr,
				VolumeMgr:     tt.fields.VolumeMgr,
				NetworkMgr:    tt.fields.NetworkMgr,
				IOs:           tt.fields.IOs,
				ExecProcesses: tt.fields.ExecProcesses,
				Config:        tt.fields.Config,
				cache:         tt.fields.cache,
				monitor:       tt.fields.monitor,
			}
			got, err := mgr.container(tt.args.nameOrPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerManager.container() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerManager.container() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainerManager_generateID(t *testing.T) {
	type fields struct {
		Store         *meta.Store
		Client        *ctrd.Client
		NameToID      *collect.SafeMap
		ImageMgr      ImageMgr
		VolumeMgr     VolumeMgr
		NetworkMgr    NetworkMgr
		IOs           *containerio.Cache
		ExecProcesses *collect.SafeMap
		Config        *config.Config
		cache         *collect.SafeMap
		monitor       *ContainerMonitor
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &ContainerManager{
				Store:         tt.fields.Store,
				Client:        tt.fields.Client,
				NameToID:      tt.fields.NameToID,
				ImageMgr:      tt.fields.ImageMgr,
				VolumeMgr:     tt.fields.VolumeMgr,
				NetworkMgr:    tt.fields.NetworkMgr,
				IOs:           tt.fields.IOs,
				ExecProcesses: tt.fields.ExecProcesses,
				Config:        tt.fields.Config,
				cache:         tt.fields.cache,
				monitor:       tt.fields.monitor,
			}
			got, err := mgr.generateID()
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerManager.generateID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ContainerManager.generateID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainerManager_generateName(t *testing.T) {
	type fields struct {
		Store         *meta.Store
		Client        *ctrd.Client
		NameToID      *collect.SafeMap
		ImageMgr      ImageMgr
		VolumeMgr     VolumeMgr
		NetworkMgr    NetworkMgr
		IOs           *containerio.Cache
		ExecProcesses *collect.SafeMap
		Config        *config.Config
		cache         *collect.SafeMap
		monitor       *ContainerMonitor
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &ContainerManager{
				Store:         tt.fields.Store,
				Client:        tt.fields.Client,
				NameToID:      tt.fields.NameToID,
				ImageMgr:      tt.fields.ImageMgr,
				VolumeMgr:     tt.fields.VolumeMgr,
				NetworkMgr:    tt.fields.NetworkMgr,
				IOs:           tt.fields.IOs,
				ExecProcesses: tt.fields.ExecProcesses,
				Config:        tt.fields.Config,
				cache:         tt.fields.cache,
				monitor:       tt.fields.monitor,
			}
			if got := mgr.generateName(tt.args.id); got != tt.want {
				t.Errorf("ContainerManager.generateName() = %v, want %v", got, tt.want)
			}
		})
	}
}
