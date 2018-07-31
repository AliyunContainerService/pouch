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
		// test case
		{
			name: "normal test",
			args: args{ports: []string{"192.168.0.1:8080:1010"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"1010/tcp": []types.PortBinding{
					types.PortBinding{HostIP: "192.168.0.1", HostPort: "8080"},
				},
			}),
			wantErr: false,
		},

		{
			name: "protocol test",
			args: args{ports: []string{"192.168.0.1:8080:1010"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"1010/udp": []types.PortBinding{
					types.PortBinding{HostIP: "192.168.0.1", HostPort: "8080"},
				},
			}),
			wantErr: false,
		},

		{
			name: "test small ip",
			args: args{ports: []string{"0.0.0.0.:8080:1010"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"1010/tcp": []types.PortBinding{
					types.PortBinding{HostIP: "0.0.0.0", HostPort: "8080"},
				},
			}),
			wantErr: false,
		},

		{
			name: "test big ip",
			args: args{ports: []string{"256.256.256.256:8080:1010"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"1010/tcp": []types.PortBinding{
					types.PortBinding{HostIP: "256.256.256.256", HostPort: "8080"},
				},
			}),
			wantErr: false,
		},
		{
			name: "test host small port",
			args: args{ports: []string{"192.168.0.1:0:1010"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"1010/tcp": []types.PortBinding{
					types.PortBinding{HostIP: "192.168.0.1", HostPort: "0"},
				},
			}),
			wantErr: false,
		},
		{
			name: "test host big port",
			args: args{ports: []string{"192.168.0.1:65536:1010"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"1010/tcp": []types.PortBinding{
					types.PortBinding{HostIP: "192.168.0.1", HostPort: "65535"},
				},
			}),
			wantErr: false,
		},

		{
			name: "test container small port",
			args: args{ports: []string{"192.168.0.1:8080:0"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"0/tcp": []types.PortBinding{
					types.PortBinding{HostIP: "192.168.0.1", HostPort: "8080"},
				},
			}),
			wantErr: false,
		},

		{
			name: "test container big port",
			args: args{ports: []string{"192.168.0.1:8080:65536"}},
			want: types.PortMap(map[string][]types.PortBinding{
				"65536/tcp": []types.PortBinding{
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
		// test case
		{
			name: "test normal port",
			args: args{
				types.PortMap(map[string][]types.PortBinding{
					"1010/tcp": []types.PortBinding{
						types.PortBinding{HostIP: "192.168.0.1", HostPort: "8080"},
					},
				}),
			},
			wantErr: false,
		},

		{
			name: "test small ip",
			args: args{
				types.PortMap(map[string][]types.PortBinding{
					"1010/tcp": []types.PortBinding{
						types.PortBinding{HostIP: "0.0.0.0", HostPort: "8080"},
					},
				}),
			},
			wantErr: false,
		},

		{
			name: "test big ip",
			args: args{
				types.PortMap(map[string][]types.PortBinding{
					"1010/tcp": []types.PortBinding{
						types.PortBinding{HostIP: "65536.65536.65536.65536", HostPort: "8080"},
					},
				}),
			},
			wantErr: false,
		},

		{
			name: "test small port",
			args: args{
				types.PortMap(map[string][]types.PortBinding{
					"1010/tcp": []types.PortBinding{
						types.PortBinding{HostIP: "192.168.0.1", HostPort: "0"},
					},
				}),
			},
			wantErr: false,
		},

		{
			name: "test big port",
			args: args{
				types.PortMap(map[string][]types.PortBinding{
					"1010/tcp": []types.PortBinding{
						types.PortBinding{HostIP: "192.168.0.1", HostPort: "65536"},
					},
				}),
			},
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
