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
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{ports: []string{"192.168.0.1:8080:1010"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"1010/tcp": []types.PortBinding{
					types.PortBinding{HostIP: "192.168.0.1", HostPort: "8080"},
				},
			}),
			wantErr: false,
		},

		{
			name: "test2",
			args: args{ports: []string{"192.168.0.1:8080:1011"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"1010/tcp": []types.PortBinding{
					types.PortBinding{HostIP: "192.168.0.1", HostPort: "8080"},
				},
			}),
			wantErr: false,
		},

		{
			name: "test3",
			args: args{ports: []string{"192.168.0.1:8090:1010"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"1010/tcp": []types.PortBinding{
					types.PortBinding{HostIP: "192.168.0.1", HostPort: "8080"},
				},
			}),
			wantErr: false,
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
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{types.PortMap(map[string][]types.PortBinding{
				"1010/tcp": []types.PortBinding{
					types.PortBinding{HostIP: "192.168.0.1", HostPort: "8080"},
				},
			})},
			wantErr: false,
		},

		{
			name: "test2",
			args: args{types.PortMap(map[string][]types.PortBinding{
				"1010": []types.PortBinding{
					types.PortBinding{HostIP: "192.168.0.1", HostPort: "8080"},
				},
			})},
			wantErr: false,
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
