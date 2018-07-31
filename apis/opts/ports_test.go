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
		//Correct test case
		//Use '-' represent multiple addresses
		//The default method is '／tcp'
		{
			name: "test1",
			args: args{
				portList: []string{"8080-8081/udp", "22"},
				expose:   []string{"8082", "22-24/tcp"},
			},
			want: map[string]interface{}{
				"8080/udp": struct{}{},
				"8081/udp": struct{}{},
				"8082/tcp": struct{}{},
				"22/tcp":   struct{}{},
				"23/tcp":   struct{}{},
				"24/tcp":   struct{}{},
			},
			wantErr: false,
		},
		//Input format error
		{
			name: "test2",
			args: args{
				portList: []string{"0.0.0.0:8080", "0.0.0.0:22"},
				expose:   []string{"0.0.0.0:8080", "0.0.0.0:22"},
			},
			want:    nil,
			wantErr: true,
		},
		//Input contains illegal characters (Chinese characters)
		{
			name: "test3",
			args: args{
				portList: []string{"8080", "8081"},
				expose:   []string{"8082／udp", "8081"},
			},
			want:    nil,
			wantErr: true,
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
