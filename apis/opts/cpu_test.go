package opts

import (
	"testing"
)

func TestValidateCPUPeriod(t *testing.T) {
	//TODO
}

func TestValidateCPUQuota(t *testing.T) {
	result := ValidateCPUQuota(0)
	if result != nil {
		t.Fatalf("Wrong return:expected nil,actual %v", result)
	}

	result := ValidateCPUQuota(1)
	if result == nil {
		t.Fatal("illegal period error is expected,but err is nil")
	}

	result := ValidateCPUQuota(1001)
	if result != nil {
		t.Fatalf("Wrong return:expected nil,actual %v", result)
	}
}
