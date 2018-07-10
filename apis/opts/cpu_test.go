package opts

import (
	"math/rand"
	"testing"
)

func TestValidateCPUPeriod(t *testing.T) {
  //TODO
}


func TestValidateCPUQuota(t *testing.T) {
	tests := []struct {
		name    string
		period  int64
		wantErr bool
	}{
		//测试用例
		{name: "test1", period: 0, wantErr: false},
		{name: "test2", period: rand.Int63n(1000), wantErr: true},
		{name: "test2", period: 1001, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCPUPeriod(tt.period)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCPUQuota() error = %v", err) //若不符合预期则报错
			}
		})
	}
}
