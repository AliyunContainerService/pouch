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
		// {
		// 	name:    "test1",
		// 	args:    args{portList: []string{"100:200"}, expose: []string{"100"}},
		// 	want:    map[string][]types.PortBinding{"200/tcp": []types.PortBinding{types.PortBinding{HostIP: "", HostPort: "100"}}},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "test2",
		// 	args:    args{portList: []string{"101:200"}, expose: []string{"101"}},
		// 	want:    map[string][]types.PortBinding{"200/tcp": []types.PortBinding{types.PortBinding{HostIP: "", HostPort: "100"}}},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "test3",
		// 	args:    args{portList: []string{"102:200"}, expose: []string{"102", "200"}},
		// 	want:    map[string][]types.PortBinding{"200/tcp": []types.PortBinding{types.PortBinding{HostIP: "127.0.0.1", HostPort: "100"}}},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "test4",
		// 	args:    args{portList: []string{"103:200"}, expose: []string{"103"}},
		// 	want:    map[string][]types.PortBinding{"200/tcp": []types.PortBinding{types.PortBinding{HostIP: "", HostPort: "103"}}},
		// 	wantErr: false,
		// },
		{
			name:    "test5",
			args:    args{portList: []string{"104:200"}, expose: []string{"104", "200"}},
			wantErr: true,
		},
		{
			name:    "test6",
			args:    args{portList: []string{"104:200"}, expose: []string{"104"}},
			wantErr: false,
		},
		{
			name:    "test7",
			args:    args{portList: []string{"104:200"}, expose: []string{"104", "200"}},
			wantErr: true,
		},
		{
			name:    "test8",
			args:    args{portList: []string{"104:200"}, expose: []string{"104", "200"}},
			wantErr: true,
		},
		{
			name:    "test9",
			args:    args{portList: []string{"105:200"}, expose: []string{"105", "200"}},
			wantErr: false,
		},
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
