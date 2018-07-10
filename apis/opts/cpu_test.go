package opts

import (
	"math/rand"
	"testing"
)

func TestValidateCPUPeriod(t *testing.T) {
	tests := []struct {
		name    string
		period  int64
		wantErr bool
	}{
		//测试用例
		{name: "test1", period: 0, wantErr: false},
		{name: "test2", period: 999, wantErr: true},
		{name: "test3", period: 1000 + rand.Int63n(999001), wantErr: false},
		{name: "test4", period: 1000001, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCPUPeriod(tt.period)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCPUPeriod() error = %v", err) //若不符合预期则报错
			}
		})
	}
}

func TestValidateCPUQuota(t *testing.T) {
  //TODO
}