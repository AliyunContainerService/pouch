package opts

import (
	"testing"
)

func TestValidateCPUPeriod(t *testing.T) {
	result := ValidateCPUPeriod(0)
	if result != nil {
		t.Fatalf("Wrong return:expected nil,actual %v", result)
	}

	result = ValidateCPUPeriod(1)
	if result == nil {
		t.Fatal("illegal period error is expected,but err is nil")
	}

	result = ValidateCPUPeriod(1000001)
	if result == nil {
		t.Fatal("illegal period error is expected,but err is nil")
	}

	result = ValidateCPUPeriod(1001)
	if result != nil {
		t.Fatalf("Wrong return:expected nil,actual %v", result)
	}
}

func TestValidateCPUQuota(t *testing.T) {
	// TODO
}
