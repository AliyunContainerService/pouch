package opts

import (
	"reflect"
	"testing"
)

func TestParseExposedPorts(t *testing.T) {
	type args struct {
		portList []string
		expose   []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
	// TODO: Add test cases.
			{
				name:"case1",
				args:args{
					portlist:[]string{
					"80",
				},
					expose:[]string{
					"127.0.0.1:80",
				},
			},
				want:nil,
				wantErr:true,
		},
		{
			name:"case2",
			args:args{
				portlist:[]string{
					"88/88",
				},
				expose:[]string{
					"127.0.0.1:80",
				},
			},
			want:nil,
			wantErr:true,
		},
		{
			name:"case3",
			args:args{
				portlist:[]string{
					"11/11:80:80",
				},
				expose:[]string{
					"127.0.0.1:80",
				},
		},
			want:nil,
			wantErr:true,
		},
		{
			name:"case4",
			args:args{
				portlist:[]string{
					"/81:80:80",
				},
				expose:[]string{
					"127.0.0.1:80",
				},
			},
			want:nil,
			wantErr:true,
		},
		{
			name:"case5",
			args:args{
				protlist:[]string{
					"88:",
				},
				expose:[]string{
					"127.0.0.1:80",
				},
			},
			want:nil,
			wantErr:true,
		},
		{
			name:"case6",
			args:args{
				portlist:[]string{
					"99:99:99",
				},
				expose:[]string{
					"127.0.0.1:80",
				},
			},
			want:nil,
			wantErr:true,
		},
		{
			name:"case7",
			args:args{
				portlist:[]string{
					"300.300.300.300:22:22",
				},
				expose:[]string{
					"127.0.0.1:80",
				},
			},
			want:nil,
			wantErr:true,
		},
		{
			name:"case8",
			args:args{
				portlist：[]string{
					"300.300.300.300:22:22",
				},
				expose:[]string{
					"127.0.0.1:80hg99",
				},
			},
			want:nil,
			wantErr:true,
		},
		{
			name:"test9",
			args:args{
				portList: nil, 
				expose: nil
				},
			want:    map[string]interface{}{},
			wantErr: false},
		{
			name: "test10",
			args: args{
				portList: []string{"168.0.0.1:80:80/tcp"},
				expose: nil
				},
			want:    map[string]interface{}{"80/tcp": struct{}{}},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExposedPorts(tt.args.portList, tt.args.expose)
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
