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
		{
			name: "getIpamConfigInputNil",
			args: args{
				data: []apitypes.IPAMConfig{},
			},
			want:    []*libnetwork.IpamConf{},
			want1:   []*libnetwork.IpamConf{},
			wantErr: false,
		},
		{
			name: "getIpamConfigInputIpv4Normal",
			args: args{
				data: []apitypes.IPAMConfig{
					{
						AuxAddress: map[string]string{
							"abc": "def",
						},
						Gateway: "192.168.1.1",
						IPRange: "192.168.1.1/26",
						Subnet:  "192.168.1.0/24",
					},
				},
			},
			want: []*libnetwork.IpamConf{
				{
					PreferredPool: "192.168.1.0/24",
					SubPool:       "192.168.1.1/26",
					Gateway:       "192.168.1.1",
					AuxAddresses: map[string]string{
						"abc": "def",
					},
				},
			},
			want1:   []*libnetwork.IpamConf{},
			wantErr: false,
		},
		{
			name: "getIpamConfigIPv6Normal",
			args: args{
				data: []apitypes.IPAMConfig{
					{
						AuxAddress: map[string]string{
							"foo": "bar",
						},
						Gateway: "2002:db8:1::1",
						IPRange: "2002:db8:1::1/68",
						Subnet:  "2002:db8:1::1/64",
					},
				},
			},
			want:    []*libnetwork.IpamConf{},
			wantErr: false,
			want1: []*libnetwork.IpamConf{
				{
					PreferredPool: "2002:db8:1::1/64",
					SubPool:       "2002:db8:1::1/68",
					Gateway:       "2002:db8:1::1",
					AuxAddresses: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		// TODO: add Invalid IPv4 subnet

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
