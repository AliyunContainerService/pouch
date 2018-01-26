package utils

import (
	"crypto/tls"
	"reflect"
	"testing"
)

func TestGenTLSConfig(t *testing.T) {
	type args struct {
		key  string
		cert string
		ca   string
	}
	tests := []struct {
		name    string
		args    args
		want    *tls.Config
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenTLSConfig(tt.args.key, tt.args.cert, tt.args.ca)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenTLSConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
