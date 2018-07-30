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
		{name: "test1", args: args{portList: nil, expose: nil}, want: map[string]interface{}{}, wantErr: false},
		{name: "test2", args: args{portList: []string{"127.0.0.1:2:2/tcp"}, expose: nil}, want: map[string]interface{}{"2/tcp": struct{}{}}, wantErr: false},
		{name: "test3", args: args{portList: nil, expose: []string{"20-21/tcp"}}, want: map[string]interface{}{"20/tcp": struct{}{}, "21/tcp":struct{}{}}, wantErr: false},
		{name: "test4", args: args{portList: []string{"127.0.0.1:2:2/tcp"}, expose: []string{"20-21/tcp"}}, want: map[string]interface{}{"2/tcp": struct{}{}, "20/tcp": struct{}{}, "21/tcp":struct{}{}}, wantErr: false},
		{name: "test5", args: args{portList: []string{"2"}, expose: nil}, want: map[string]interface{}{"2/tcp": struct{}{}}, wantErr: false},
		{name: "test6", args: args{portList: []string{"2/udp"}, expose: nil}, want: map[string]interface{}{"2/udp": struct{}{}}, wantErr: false},
		{name: "test7", args: args{portList: []string{"2/tcp"}, expose: nil}, want: map[string]interface{}{"2/tcp": struct{}{}}, wantErr: false},
		{name: "test8", args: args{portList: nil, expose: []string{"20"}}, want: map[string]interface{}{"20/tcp": struct{}{}}, wantErr: false},
		{name: "test9", args: args{portList: []string{"127.0.0.1:2:2/tcp"}, expose: []string{"20/tcp"}}, want: map[string]interface{}{"2/tcp": struct{}{}, "20/tcp": struct{}{}}, wantErr: false},	
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
