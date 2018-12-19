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
		{"test single env", args{env: []string{"foo=bar"}}, map[string]string{"foo": "bar"}, false},
		{"test error env format", args{env: []string{"ErrorInfo"}}, nil, true},
		{"test multiple envs", args{env: []string{"foo=bar", "A=a"}}, map[string]string{"foo": "bar", "A": "a"}, false},
		{"test multiple '=' envs", args{env: []string{"A=1=2"}}, map[string]string{"A": "1=2"}, false},
		{"test nil env", args{env: nil}, map[string]string{}, false}, // empty map
		{"test empty env", args{env: []string{""}}, nil, true},
		{"test empty blank", args{env: []string{"  "}}, nil, true},
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

func TestValidateEnv(t *testing.T) {
	type args struct {
		env map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test env nil", args{env: map[string]string{}}, true},
		{"test env empty map", args{env: map[string]string{}}, true},
		{"test single env map", args{env: map[string]string{"foo": "bar"}}, false},
		{"test multiple env map", args{env: map[string]string{"foo": "bar", "A": "1=2"}}, false},
		// TODO more
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnv(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
