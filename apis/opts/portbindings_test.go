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
			name:    "test1",
			args:    args{ports: []string{"100:200"}},
			want:    types.PortMap{"200/tcp": []types.PortBinding{types.PortBinding{HostIP: "", HostPort: "100"}}},
			wantErr: false,
		},
		{
			name: "test4",
			args: args{ports: []string{"100-101:200-201"}},
			want: types.PortMap{"200/tcp": []types.PortBinding{types.PortBinding{HostIP: "", HostPort: "100"}},
				"201/tcp": []types.PortBinding{types.PortBinding{HostIP: "", HostPort: "101"}}},
			wantErr: false,
		},
		{
			name:    "test2",
			args:    args{ports: []string{"127.0.0.1:100:200/UDP"}},
			want:    types.PortMap{"200/udp": []types.PortBinding{types.PortBinding{HostIP: "127.0.0.1", HostPort: "100"}}},
			wantErr: false,
		},
		{
			name:    "test3",
			args:    args{ports: []string{"127.0.0.1:100ss:200/UDP"}},
			wantErr: true,
		},
		{
			name:    "test5",
			args:    args{ports: []string{""}},
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
			name: "test1",
			args: args{
				portBindings: map[string][]types.PortBinding{
					"8181/tcp": {types.PortBinding{
						"127.0.0.1", "8080",
					}},
				},
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				portBindings: map[string][]types.PortBinding{
					"818ss1": {types.PortBinding{
						"127.0.0.1", "8080",
					}},
				},
			},
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
