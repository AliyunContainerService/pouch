package opts

import (
	"testing"
)

func TestValidateCPUPeriod(t *testing.T) {

	tests := []struct {
		name    string
		period  int64
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", period: 0, wantErr: false},
		{name: "test2", period: 10, wantErr: true},
		{name: "test3", period: 2000000, wantErr: true},
		{name: "test4", period: 1000, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCPUPeriod(tt.period); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCPUPeriod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	// TODO
}

func TestValidateCPUQuota(t *testing.T) {
	// TODO
}
