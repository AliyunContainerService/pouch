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
		{name: "test1", args: args{ports: []string{"100:200"}}, want: types.PortMap{"200/tcp": []types.PortBinding{{HostIP: "", HostPort: "100"}}}, wantErr: false},
		{name: "test2", args: args{ports: []string{"127.0.0.1:100:200/UDP"}}, want: types.PortMap{"200/udp": []types.PortBinding{{HostIP: "127.0.0.1", HostPort: "100"}}}, wantErr: false},
		{name: "test3", args: args{ports: []string{"127.0.0.1:100-101:200/UDP"}}, want: types.PortMap{"200/udp": []types.PortBinding{{HostIP: "127.0.0.1", HostPort: "100"}}}, wantErr: true},
		{name: "test4", args: args{ports: []string{"127.0.0.1:100-101:200-201/UDP"}}, want: types.PortMap{"200/udp": []types.PortBinding{{HostIP: "127.0.0.1", HostPort: "100"}}, "201/udp": []types.PortBinding{{HostIP: "127.0.0.1", HostPort: "101"}}}, wantErr: false},
		{name: "test5", args: args{ports: []string{"127.0.0.1:100:65536/UDP"}}, want: types.PortMap{"65536/udp": []types.PortBinding{{HostIP: "127.0.0.1", HostPort: "100"}}}, wantErr: true},
		{name: "test6", args: args{ports: []string{"127.0.0.1:100:65534/scc"}}, want: types.PortMap{"65534/udp": []types.PortBinding{{HostIP: "127.0.0.1", HostPort: "100"}}}, wantErr: true},
		{name: "test7", args: args{ports: []string{"[FF01::1101]:100:200/sctp"}}, want: types.PortMap{"200/sctp": []types.PortBinding{{HostIP: "[FF01::1101]", HostPort: "100"}}}, wantErr: false},
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
		{name: "test1", args: args{portBindings: types.PortMap{"100": []types.PortBinding{{HostIP: "", HostPort: "100"}}}}, wantErr: false},
		{name: "test2", args: args{portBindings: types.PortMap{"-2": []types.PortBinding{{HostIP: "", HostPort: "100"}}}}, wantErr: true},
		{name: "test3", args: args{portBindings: types.PortMap{"200s": []types.PortBinding{{HostIP: "", HostPort: "100"}}}}, wantErr: true},
		{name: "test4", args: args{portBindings: types.PortMap{"200": []types.PortBinding{{HostIP: "", HostPort: "100-200"}}}}, wantErr: false},
		{name: "test5", args: args{portBindings: types.PortMap{"200": []types.PortBinding{{HostIP: "", HostPort: "300-200"}}}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePortBinding(tt.args.portBindings); (err != nil) != tt.wantErr {
				t.Errorf("VerifyPortBinding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
