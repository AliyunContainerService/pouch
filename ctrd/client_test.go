package ctrd

import (
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	type args struct {
		homeDir string
		opts    []ClientOpt
	}
	tests := []struct {
		name    string
		args    args
		want    APIClient
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.homeDir, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
