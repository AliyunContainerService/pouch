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
		{name: "case 1",
			args:    args{ports: []string{"192.168.1.1:8888:20/tcp", "192.168.1.1:9999:30/tcp"}},
			want:    types.PortMap{"20/tcp": {{"192.168.1.1", "8888"}}, "30/tcp": {{"192.168.1.1", "9999"}}},
			wantErr: false},
		{name: "test1",
			args:    args{ports: []string{"127.0.0.1:1111:1234", "127.0.0.2:2222:1234", "127.0.0.2:5555:8080"}},
			want:    types.PortMap{"1234/tcp": []types.PortBinding{{"127.0.0.1", "1111"}, {"127.0.0.2", "2222"}}, "8080/tcp": []types.PortBinding{{"127.0.0.2", "5555"}}},
			wantErr: false},
		{name: "test2",
			args:    args{[]string{"127.0.0.1:1234", "127.0.0.2:2222:1234"}},
			want:    nil,
			wantErr: true},
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
		{"test1", args{types.PortMap{"1234/tcp": []types.PortBinding{types.PortBinding{HostIP: "127.0.0.1", HostPort: "1111"}, types.PortBinding{HostIP: "127.0.0.2", HostPort: "2222"}}, "8080/tcp": []types.PortBinding{types.PortBinding{HostIP: "127.0.0.1", HostPort: "5555"}}}}, false},
		{"test2", args{types.PortMap{"1234tcp": []types.PortBinding{}}}, true},
		{"test3", args{types.PortMap{"1234/tcp": []types.PortBinding{types.PortBinding{HostIP: "127.0.0.1", HostPort: "abc/1111"}}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePortBinding(tt.args.portBindings); (err != nil) != tt.wantErr {
				t.Errorf("VerifyPortBinding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
