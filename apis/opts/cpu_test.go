package opts

import (
	"math/rand"
	"testing"
	"time"
)

func TestValidateCPUPeriod(t *testing.T) {
	rand.Seed(time.Now().Unix())

	err := ValidateCPUPeriod(0)
	if err != nil {
		t.Fatalf("validate cpu peroid 0 err: %v", err)
	}

	p := rand.Int63n(999001) + 1000
	err = ValidateCPUPeriod(p)
	if err != nil {
		t.Fatalf("validate cpu peroid %v err: %v", p, err)
	}

	err1 := ValidateCPUPeriod(999)
	err2 := ValidateCPUPeriod(1000001)
	if err1 == nil || err2 == nil {
		t.Fatalf("expect validate cpu peroid error, but err is nil")
	}
}

func TestValidateCPUQuota(t *testing.T) {
	// TODO
}
