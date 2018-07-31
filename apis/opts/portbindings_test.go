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
			name: "normal successful cases",
			args: args{
				ports: []string{
					"10.3.8.211:211:21/tcp",
					"[fe80::1464:8a92:d6a:a73b]:211:212/tcp",
				},
			},
			want: map[string][]types.PortBinding{
				"212/tcp": []types.PortBinding{
					{
						HostIP:   "10.3.8.21",
						HostPort: "211",
					},
					{
						HostIP:   "fe80::1464:8a92:d6a:a73b",
						HostPort: "211",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "normal case with empty string",
			args: args{
				ports: []string{},
			},
			want:    map[string][]types.PortBinding{},
			wantErr: false,
		},
		{
			name: "failure case with Invalid ip address",
			args: args{
				ports: []string{
					"10.3.8.256:211:212/tcp",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "failure case with wrong proto",
			args: args{
				ports: []string{
					"10.3.8.256:211:212/ftp",
				},
			},
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
			name: "successful cases",
			args: args{portBindings: types.PortMap{
				"23232/tcp": []types.PortBinding{
					{HostIP: "10.3.8.211", HostPort: "212"}},
				"23232": []types.PortBinding{
					{HostIP: "10.3.8.211", HostPort: "212"}},
			}},
			wantErr: false,
		},
		{
			name: "invalidSplitCharacter",
			args: args{portBindings: types.PortMap{
				"23232:tcp": []types.PortBinding{
					{HostIP: "10.3.8.211", HostPort: "212"}},
			}},
			wantErr: true,
		},
		{
			name: "invalidPort",
			args: args{portBindings: types.PortMap{
				"65536/tcp": []types.PortBinding{
					{HostIP: "10.3.8.211", HostPort: "212"}},
			}},
			wantErr: true,
		},
		{
			name: "invalidHostPort",
			args: args{portBindings: types.PortMap{
				"23232/tcp": []types.PortBinding{
					{HostIP: "10.3.8.211", HostPort: "fd"}},
			}},
			wantErr: true,
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
