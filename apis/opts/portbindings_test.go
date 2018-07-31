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
			name: "normal test",
			args: args{ports: []string{"112.126.1.1:1201-1203:1301-1303/tcp"}},
			want: types.PortMap{
				"1301/tcp": []types.PortBinding{
					types.PortBinding{
						HostIP:   "112.126.1.1",
						HostPort: "1201",
					},
				},
				"1302/tcp": []types.PortBinding{
					types.PortBinding{
						HostIP:   "112.126.1.1",
						HostPort: "1202",
					},
				},
				"1303/tcp": []types.PortBinding{
					types.PortBinding{
						HostIP:   "112.126.1.1",
						HostPort: "1203",
					},
				},
			},
			wantErr: false,
		},

		{
			name: "normal test",
			args: args{ports: []string{"112.126.1.1:1201:1301"}},
			want: types.PortMap{
				"1301/tcp": []types.PortBinding{
					types.PortBinding{
						HostIP:   "112.126.1.1",
						HostPort: "1201",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "fail test",
			args:    args{ports: []string{"112.126.1.1:1201:/tcp"}},
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
			name: "normal test",
			args: args{types.PortMap{
				"1301/tcp": []types.PortBinding{
					types.PortBinding{
						HostIP:   "112.126.1.1",
						HostPort: "1205/udp",
					},
				},
				"1302/tcp": []types.PortBinding{
					types.PortBinding{
						HostIP:   "112.126.1.1",
						HostPort: "1202",
					},
				},
				"1303/tcp": []types.PortBinding{
					types.PortBinding{
						HostIP:   "112.126.1.1",
						HostPort: "1203",
					},
				},
			}},
			wantErr: false,
		},
		{
			name: "normal test",
			args: args{types.PortMap{
				"1301": []types.PortBinding{
					types.PortBinding{
						HostIP:   "112.126.1.1",
						HostPort: "1205-1207",
					},
				},
				"1302": []types.PortBinding{
					types.PortBinding{
						HostIP:   "112.126.1.1",
						HostPort: "1202",
					},
				},
			}},
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
