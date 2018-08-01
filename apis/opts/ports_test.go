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
		{name: "test1", args: args{portList: []string{""}, expose: []string{""}}, want: nil, wantErr: true},
		{name: "test2", args: args{portList: []string{""}, expose: []string{"1500/tcp"}}, want: nil, wantErr: true},
		{name: "test3", args: args{portList: []string{""}, expose: []string{"1500-1505"}}, want: nil, wantErr: true},
		{name: "test4", args: args{portList: []string{""}, expose: []string{"1500-1505/tcp"}}, want: nil, wantErr: true},
		{name: "test5", args: args{portList: []string{"1000"}, expose: []string{""}}, want: nil, wantErr: true},
		{name: "test6", args: args{portList: []string{"1000"}, expose: []string{"999-1001"}}, want: map[string]interface{}{"999/tcp": struct{}{}, "1000/tcp": struct{}{}, "1001/tcp": struct{}{}}, wantErr: false},
		{name: "test7", args: args{portList: []string{"1000", "2000"}, expose: []string{"1500/tcp"}}, want: map[string]interface{}{"1000/tcp": struct{}{}, "2000/tcp": struct{}{}, "1500/tcp": struct{}{}}, wantErr: false},
		{name: "test8", args: args{portList: []string{"1000", "2000"}, expose: []string{"1500/tcp", "1505"}}, want: map[string]interface{}{"1000/tcp": struct{}{}, "1500/tcp": struct{}{}, "1505/tcp": struct{}{}, "2000/tcp": struct{}{}}, wantErr: false},
		{name: "test9", args: args{portList: []string{"1000", "2000"}, expose: []string{"999-1001/tcp", "1505"}}, want: map[string]interface{}{"999/tcp": struct{}{}, "1000/tcp": struct{}{}, "1001/tcp": struct{}{}, "1505/tcp": struct{}{}, "2000/tcp": struct{}{}}, wantErr: false},
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
