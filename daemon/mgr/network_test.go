package mgr

import (
	"reflect"
	"testing"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/docker/libnetwork"
)

func Test_getIpamConfig(t *testing.T) {
	type args struct {
		data []apitypes.IPAMConfig
	}
	tests := []struct {
		name    string
		args    args
		want    []*libnetwork.IpamConf
		want1   []*libnetwork.IpamConf
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getIpamConfig(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIpamConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getIpamConfig() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getIpamConfig() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
