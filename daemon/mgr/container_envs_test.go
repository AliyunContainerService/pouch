package mgr

import (
	"testing"
)

func Test_updateContainerEnv(t *testing.T) {
	type args struct {
		inputRawEnv []string
		baseFs      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateContainerEnv(tt.args.inputRawEnv, tt.args.baseFs); (err != nil) != tt.wantErr {
				t.Errorf("updateContainerEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
