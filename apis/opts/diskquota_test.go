package opts

import (
	"reflect"
	"testing"
)

func TestParseDiskQuota(t *testing.T) {
	type args struct {
		diskquota []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{diskquota: []string{""}}, want: nil, wantErr: true},
		{name: "test2", args: args{diskquota: []string{"foo=foo=foo"}}, want: nil, wantErr: true},
		{name: "test3", args: args{diskquota: []string{"foo"}}, want: map[string]string{"/": "foo"}, wantErr: false},
		{name: "test4", args: args{diskquota: []string{"foo=foo"}}, want: map[string]string{"foo": "foo"}, wantErr: false},
		{name: "test5", args: args{diskquota: []string{"foo=foo", "bar=bar"}}, want: map[string]string{"foo": "foo", "bar": "bar"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDiskQuota(tt.args.diskquota)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDiskQuota() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDiskQuota() = %v, want %v", got, tt.want)
			}
		})
	}
}
