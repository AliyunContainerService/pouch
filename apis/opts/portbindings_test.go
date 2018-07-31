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
		//good case
		{	name:	"case 1",
			args:	args{ports:	[]string{"192.168.1.1:8888:20/tcp", "192.168.1.1:9999:30/tcp"}},
			want:	types.PortMap{"20/tcp": {{"192.168.1.1", "8888"}}, "30/tcp": {{"192.168.1.1", "9999"}},},
			wantErr:	false,
		},
		//bad case: ip
		{
			name:	"case bad 1",
			args:	args{ports:	[]string{"badIp:8888:20/tcp"}},
			want:	nil,
			wantErr:	true,
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
		//good case
		{
			name:	"case verify 1",
			args:	args{portBindings:	types.PortMap{"20/tcp": {{"192.168.1.1", "8888"}}},},
			wantErr:	false,
		},
		//bad case: bad cantainer port
		{
			name:	"bad case verify 1",
			args:	args{portBindings:	types.PortMap{"badPort/tcp": {{"192.168.1.1", "8888"}}},},
			wantErr:	true,
		},
		//bad case: bad host port
		{
			name:	"bad case verify 2",
			args:	args{portBindings:      types.PortMap{"20/tcp": {{"192.168.1.1", "badPort"}}},},
			wantErr:	true,
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
