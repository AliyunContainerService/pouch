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
		{name: "test 1", args: args{portBindings: map[string][]types.PortBinding{"21/ftp": {types.PortBinding{HostIP: "", HostPort: "21"}}}}, wantErr: false},
		{name: "test 2", args: args{portBindings: map[string][]types.PortBinding{"21/udp": {types.PortBinding{HostIP: "", HostPort: "21"}}}}, wantErr: false},
		{name: "test 3", args: args{portBindings: map[string][]types.PortBinding{"21/ftp": {types.PortBinding{HostIP: "", HostPort: "35"}}}}, wantErr: false},
		{name: "test 4", args: args{portBindings: map[string][]types.PortBinding{"21/ftp": {types.PortBinding{HostIP: "0.0.0.0", HostPort: ""}}}}, wantErr: false},
		{name: "test 5", args: args{portBindings: map[string][]types.PortBinding{"65537/tcp": {types.PortBinding{HostIP: "", HostPort: ""}}}}, wantErr: true},
		{name: "test 6", args: args{portBindings: map[string][]types.PortBinding{"0/tcp": {types.PortBinding{HostIP: "", HostPort: ""}}}}, wantErr: false},
		{name: "test 7", args: args{portBindings: map[string][]types.PortBinding{"0/tcp": {types.PortBinding{HostIP: "0.0.0.0", HostPort: "0"}}}}, wantErr: false},
		{name: "test 8", args: args{portBindings: map[string][]types.PortBinding{"80/http": {types.PortBinding{HostIP: "", HostPort: ""}}}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePortBinding(tt.args.portBindings); (err != nil) != tt.wantErr {
				t.Errorf("VerifyPortBinding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
