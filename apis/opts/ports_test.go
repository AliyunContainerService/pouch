package opts

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseExposedPorts(t *testing.T) {
	type args struct {
		portList []string
		expose   []string
	}

	type TestCase struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}
	w := make(map[string]interface{})
	w["1/tcp"] = struct{}{}
	w["65535/tcp"] = struct{}{}
	w["80/tcp"] = struct{}{}

	w3 := make(map[string]interface{})
	w3["53/tcp"] = struct{}{}

	w5 := make(map[string]interface{})
	w5["53/tcp"] = struct{}{}
	w5["20/tcp"] = struct{}{}
	w5["21/tcp"] = struct{}{}
	w5["22/tcp"] = struct{}{}
	tests := []TestCase{
		TestCase{
			name: "test1",
			args: args{
				portList: []string{
					"1", "65535",
				},
				expose: []string{
					"80/tcp",
				},
			},
			want:    w,
			wantErr: false,
		},
		TestCase{
			name: "test2",
			args: args{
				portList: []string{
					"0", "65536",
				},
				expose: []string{},
			},
			want:    nil,
			wantErr: true,
		},
		TestCase{
			name: "test21",
			args: args{
				portList: []string{},
				expose: []string{
					"-1/tcp",
				},
			},
			want:    nil,
			wantErr: true,
		},
		TestCase{
			name: "test3",
			args: args{
				portList: []string{
					"53",
				},
				expose: []string{},
			},
			want:    w3,
			wantErr: false,
		},
		TestCase{
			name: "test4",
			args: args{
				portList: nil,
				expose:   nil,
			},
			want:    make(map[string]interface{}),
			wantErr: false,
		},
		TestCase{
			name: "test5",
			args: args{
				portList: []string{
					"53",
				},
				expose: []string{
					"20-22/tcp",
				},
			},
			want:    w5,
			wantErr: false,
		},
		TestCase{
			name: "test6",
			args: args{
				portList: []string{},
				expose: []string{
					"24~22/tcp",
				},
			},
			want:    nil,
			wantErr: true,
		},
		TestCase{
			name: "test7",
			args: args{
				portList: []string{},
				expose: []string{
					"20~22/udpsa",
				},
			},
			want:    nil,
			wantErr: true,
		},
		TestCase{
			name: "test8",
			args: args{
				portList: []string{},
				expose:   []string{},
			},
			want:    make(map[string]interface{}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExposedPorts(tt.args.portList, tt.args.expose)
			fmt.Printf("%+v\n", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExposedPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseExposedPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}
