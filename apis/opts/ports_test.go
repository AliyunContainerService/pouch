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
				name:	"wrongcase",
				args:	args{portList: []string{"2222"}, expose: []string{"4444"}},
				want:	map[string]interface{}{"2222/tcp":""},
				wantErr:true,
		},

		{
				name:	"wrongcase1",
				args:	args{portList: []string{"111a"}, expose: []string{"4444"}},
				want:	nil,
				wantErr:true,
		},

		{
				name:	"wrongcase2",
				args:	args{portList: []string{"1111","2222","3333"}, expose: []string{"4:aa"}},
				want:	nil,
				wantErr:true,
		},

		{
				name:	"wrongcase3",
				args:	args{portList: []string{"1111"}, expose: []string{"4444-4440"}},
				want: nil,
				wantErr:true,
		},

		{
				name:	"case3",
				args:	args{portList: []string{"120.202.45.45:2222:3333/tcp"}, expose: []string{"44aa"}},
				want: nil,
				wantErr:true,
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
