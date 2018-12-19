package opts

import (
	"reflect"
	"testing"

	"github.com/alibaba/pouch/apis/types"
)

func TestParsePortBinding(t *testing.T) {
	type args struct {
		ports []string
	}
	tests := []struct {
		name    string
		args    args
		want    types.PortMap
		wantErr bool
	}{
		{
			name: "testIpPrivatePublicBinding",
			args: args{ports: []string{"127.0.0.1:80:80"}},
			want: map[string][]types.PortBinding{
				"80/tcp": {
					{
						HostIP:   "127.0.0.1",
						HostPort: "80",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "testIpPrivatePublicProtoBinding",
			args: args{ports: []string{"127.0.0.1:80:80/udp"}},
			want: map[string][]types.PortBinding{
				"80/udp": {
					{
						HostIP:   "127.0.0.1",
						HostPort: "80",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "testPrivatePublicBinding",
			args: args{ports: []string{"26:22"}},
			want: map[string][]types.PortBinding{
				"22/tcp": {
					{
						HostIP:   "",
						HostPort: "26",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "testUnExistedProtoBinding",
			args:    args{ports: []string{"80:80/abc"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "testIPv4AddressFormatErrorBinding",
			args:    args{ports: []string{"256.0.0.1:80:80"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "testPrivatePortOutOfRange",
			args:    args{ports: []string{"65537:22"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "testPublicPortOutOfRange",
			args:    args{ports: []string{"22:65537"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePortBinding(tt.args.ports)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePortBinding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePortBinding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyPortBinding(t *testing.T) {
	type args struct {
		portBindings types.PortMap
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "testFtpPortBinding",
			args: args{
				portBindings: map[string][]types.PortBinding{
					"21/ftp": {
						types.PortBinding{HostIP: "", HostPort: "21"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "testContainerPortOutOfRange",
			args: args{
				portBindings: map[string][]types.PortBinding{
					"65537/tcp": {
						types.PortBinding{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "testZeroContainerPort",
			args: args{
				portBindings: map[string][]types.PortBinding{
					"0/tcp": {
						types.PortBinding{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "testHttpPortBinding",
			args: args{
				portBindings: map[string][]types.PortBinding{
					"80/http": {
						types.PortBinding{},
					},
				},
			}, wantErr: false,
		},
		{
			name: "testHostPortOutOfRange",
			args: args{
				portBindings: map[string][]types.PortBinding{
					"80/http": {
						types.PortBinding{HostIP: "", HostPort: "65537"},
					},
				},
			}, wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePortBinding(tt.args.portBindings); (err != nil) != tt.wantErr {
				t.Errorf("VerifyPortBinding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
