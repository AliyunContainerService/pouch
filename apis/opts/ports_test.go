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
			name:    "test1",
			args:    args{portList: []string{"0.0.0.0:443:443/tcp"}, expose: nil},
			want:    map[string]interface{}{"443/tcp": struct{}{}},
			wantErr: false,
		},
		{
			name:    "test2",
			args:    args{portList: []string{"80/tcp"}, expose: []string{"80"}},
			want:    map[string]interface{}{"80/tcp": struct{}{}},
			wantErr: false,
		},
		{
			name:    "test3",
			args:    args{portList: []string{"0.0.0.0:65535:65535/udp"}, expose: []string{"65535/udp"}},
			want:    map[string]interface{}{"65535/udp": struct{}{}},
			wantErr: false,
		},
		{
			name: "test4",
			args: args{portList: []string{"0.0.0.0:80:80/udp"}, expose: []string{"78-81/udp"}},
			want: map[string]interface{}{
				"78/udp": struct{}{},
				"79/udp": struct{}{},
				"80/udp": struct{}{},
				"81/udp": struct{}{},
			},
			wantErr: false,
		},
		{
			name:    "test5",
			args:    args{portList: nil, expose: nil},
			want:    map[string]interface{}{},
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
