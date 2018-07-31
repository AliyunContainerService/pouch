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
	type TestCase struct {
		name    string
		args    args
		want    types.PortMap
		wantErr bool
	}
	tests := []TestCase{
	  TestCase{
	  	name: "test_tcp",
	  	args: args{[]string{
	  		"127.0.0.1:3306:3306/tcp",
		}},
	  	want: types.PortMap{
	  		"3306/tcp":[]types.PortBinding{types.PortBinding{HostIP:"127.0.0.1", HostPort:"3306"}},
	  },
	  	//wantErr: true,
	  },
		TestCase{
			name: "test_udp",
			args: args{[]string{
				"127.0.0.1:3306:3306/udp",
			}},
			want: types.PortMap{
				"3306/udp":[]types.PortBinding{types.PortBinding{HostIP:"127.0.0.1", HostPort:"3306"}},
			},
			//wantErr: true,
		},
		TestCase{
			name: "test_kong",
			args: args{[]string{
				"127.0.0.1:3306:3306/",
			}},
			want: types.PortMap{
				"3306/":[]types.PortBinding{types.PortBinding{HostIP:"127.0.0.1", HostPort:"3306"}},
			},
			//wantErr: false,
		},
		TestCase{
			name: "test_valid_proto",
			args: args{[]string{
				"127.0.0.1:3306:3306/uup",
			}},
			want: types.PortMap{
				"3306/uup":[]types.PortBinding{types.PortBinding{HostIP:"127.0.0.1", HostPort:"3306"}},
			},
			//wantErr: false,
		},
		TestCase{
			name: "test_invalid_hostport",
			args: args{[]string{
				"127.0.0.1:3306/udp",
			}},
			want: types.PortMap{
				"3306/uup":[]types.PortBinding{types.PortBinding{HostIP:"127.0.0.1", HostPort:"3306"}},
			},
			//wantErr: false,
		},
		TestCase{
			name: "test_invalid_hostport",
			args: args{[]string{
				"127.0.0.1:3306/udp",
			}},
			want: types.PortMap{
				"3306/uup":[]types.PortBinding{types.PortBinding{HostIP:"127.0.0.1", HostPort:"3306"}},
			},
			//wantErr: false,
		},
		TestCase{
			name: "test_portempty",
			args: args{[]string{
				"",/Users/alice/go/src/github.com/alibaba/pouch
			}},
			want: types.PortMap{
				"3306/tcp":[]types.PortBinding{types.PortBinding{HostIP:"127.0.0.1", HostPort:"3306"}},
			},
			//wantErr: false,
		},
		TestCase{
			name: "test_0",
			args: args{[]string{
				"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			}},
			want: types.PortMap{
				"3306/tcp":[]types.PortBinding{types.PortBinding{HostIP:"127.0.0.1", HostPort:"3306"}},
			},
			//wantErr: false,
		},
		TestCase{
			name: "test_Invalid_containerPort",
			args: args{[]string{
				"127.0.0.1:330699/udp",
			}},
			want: types.PortMap{
				"3306/tcp":[]types.PortBinding{types.PortBinding{HostIP:"127.0.0.1", HostPort:"3306"}},
			},
			//wantErr: false,
		},
		TestCase{
			name: "test_Invalid_containerPort/",
			args: args{[]string{
				"127.0.0.1:3306udp",
			}},
			want: types.PortMap{
				"3306/tcp":[]types.PortBinding{types.PortBinding{HostIP:"127.0.0.1", HostPort:"3306"}},
			},
			//wantErr: false,
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
		//{
		//	name: "test3",
		//	args: args{types.PortMap{
		//		"3306": []types.PortBinding{ {HostPort:"3306",HostIP: "127.0.0.1"} },
		//	}},
		//	//wantErr: true,
		//},
		{
			name: "test_invalud_port_specification",
			args: args{types.PortMap{
				"3306": []types.PortBinding{ {HostPort:"3333306",HostIP: "127.0.0.1"} },
			}},
			//wantErr: true,
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
