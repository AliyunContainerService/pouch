package opts

import (
	"reflect"
	"testing"
)

func TestParseAnnotation(t *testing.T) {
	type args struct {
		annotations []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{annotations: []string{""}}, want: nil, wantErr: true},
		{name: "test2", args: args{annotations: []string{"=foo"}}, want: nil, wantErr: true},
		{name: "test3", args: args{annotations: []string{"key="}}, want: nil, wantErr: true},
		{name: "test4", args: args{annotations: []string{"key=foo=bar"}}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAnnotation(tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAnnotation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAnnotation() = %v, want %v", got, tt.want)
			}
		})
	}
}
