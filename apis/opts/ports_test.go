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
		{
			name: "normal test",
			args: args{
				portList: []string{"192.168.0.1:70:70/tcp"},
				expose: []string{"80/tcp"},
			},
			want: map[string]interface{}{
				"70/tcp":struct{}{},
				"80/tcp":struct{}{},
			},
			wantErr: false,
		},
		{
			name: "normalRange test",
			args: args{
				portList: []string{"192.168.0.1:70-71:100-101/tcp"},
				expose: []string{"70-71/tcp"},
			},
			want: map[string]interface{}{
				"70/tcp":struct{}{},
				"71/tcp":struct{}{},
				"100/tcp":struct{}{},
				"101/tcp":struct{}{},
			},
			wantErr: false,
		},
		{
			name: "nil test",
			args: args{
				portList: nil,
				expose: nil,
			},
			want: map[string]interface{}{
			},
			wantErr: false,
		},
		{
			name: "expose nil test",
			args: args{
				portList: []string{"192.168.0.1:70:70/tcp"},
				expose: nil,
			},
			want: map[string]interface{}{
				"70/tcp":struct{}{},
			},
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
