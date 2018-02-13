package main

import (
	"testing"

	"github.com/alibaba/pouch/apis/types"
)

func Test_parseVolume(t *testing.T) {
	type args struct {
		volumeCreateConfig *types.VolumeCreateConfig
		v                  *VolumeCreateCommand
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "parseVolumeTestOk",
			args: args{
				volumeCreateConfig: &types.VolumeCreateConfig{
					DriverOpts: map[string]string{},
					Labels:     map[string]string{},
				},
				v: &VolumeCreateCommand{
					labels:    []string{"a=b"},
					options:   []string{"1=2"},
					selectors: []string{"c=d"},
				},
			},
			wantErr: false,
		},
		{
			name: "parseVolumeTestInvalid1",
			args: args{
				volumeCreateConfig: &types.VolumeCreateConfig{
					DriverOpts: map[string]string{},
					Labels:     map[string]string{},
				},
				v: &VolumeCreateCommand{
					labels: []string{"ab"},
				},
			},
			wantErr: true,
		},
		{
			name: "parseVolumeTestInvalid2",
			args: args{
				volumeCreateConfig: &types.VolumeCreateConfig{
					DriverOpts: map[string]string{},
					Labels:     map[string]string{},
				},
				v: &VolumeCreateCommand{
					options: []string{"1&2"},
				},
			},
			wantErr: true,
		},
		{
			name: "parseVolumeTestInvalid3",
			args: args{
				volumeCreateConfig: &types.VolumeCreateConfig{
					DriverOpts: map[string]string{},
					Labels:     map[string]string{},
				},
				v: &VolumeCreateCommand{
					labels:    []string{"ab"},
					options:   []string{"1&2"},
					selectors: []string{"c==d"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseVolume(tt.args.volumeCreateConfig, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("parseVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
