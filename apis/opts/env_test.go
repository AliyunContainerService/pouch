package opts

import (
	"reflect"
	"testing"
)

func TestParseEnv(t *testing.T) {
	type args struct {
		env []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{env: []string{"foo=bar"}}, want: map[string]string{"foo": "bar"}, wantErr: false},
		{name: "test2", args: args{env: []string{"ErrorInfo"}}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEnv(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
