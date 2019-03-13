package opts

import (
	"reflect"
	"testing"
)

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{"test single env", "foo=bar", "foo=bar", false},
		{"test str with no =", "NOEQUALMARK", "NOEQUALMARK", false},
		{"test multiple '=' envs", "A=1=2", "A=1=2", false},
		{"test empty blank in value", "A=B C", "A=B C", false},
		{"test only =", "=", "", true},
		{"test empty env", "", "", true}, // empty map
		{"test empty blank", "  ", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEnv(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
