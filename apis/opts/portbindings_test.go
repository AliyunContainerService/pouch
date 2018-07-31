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
		{name: "ParsePortBinding test 1", args: args{ports: []string{"21"}}, want: map[string][]types.PortBinding{"21/tcp": {types.PortBinding{}}}, wantErr: false},
		{name: "ParsePortBinding test 2", args: args{ports: []string{"127.0.0.1:21:21"}}, want: map[string][]types.PortBinding{"21/tcp": {types.PortBinding{HostIP: "127.0.0.1", HostPort: "21"}}}, wantErr: false},
		{name: "ParsePortBinding test 3", args: args{ports: []string{"127.0.0.1:21"}}, want: nil, wantErr: true},
		{name: "ParsePortBinding test 4", args: args{ports: []string{"21:21"}}, want: map[string][]types.PortBinding{"21/tcp": {types.PortBinding{HostPort: "21"}}}, wantErr: false},
		{name: "ParsePortBinding test 2", args: args{ports: []string{"65536"}}, want: nil, wantErr: true},
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
		{name: "PortMap struct correctness test 1", args: args{portBindings: map[string][]types.PortBinding{"21/ftp": {types.PortBinding{HostIP: "", HostPort: "21"}}}}, wantErr: false},
		{name: "PortMap struct correctness test 2", args: args{portBindings: map[string][]types.PortBinding{"65537/tcp": {types.PortBinding{}}}}, wantErr: true},
		{name: "PortMap struct correctness test 3", args: args{portBindings: map[string][]types.PortBinding{"0/tcp": {types.PortBinding{}}}}, wantErr: false},
		{name: "PortMap struct correctness test 5", args: args{portBindings: map[string][]types.PortBinding{"80/http": {types.PortBinding{}}}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePortBinding(tt.args.portBindings); (err != nil) != tt.wantErr {
				t.Errorf("VerifyPortBinding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
