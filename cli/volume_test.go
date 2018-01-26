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
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseVolume(tt.args.volumeCreateConfig, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("parseVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
