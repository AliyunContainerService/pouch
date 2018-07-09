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

	for i := 0; i < 100; i++ {
		p := rand.Int63n(10000000)
		p -= 10000
		if 1000 <= p && p <= 1000000 {
			err = ValidateCPUPeriod(p)
			if err != nil {
				t.Fatalf("validate cpu peroid %v err: %v", p, err)
			}
		} else {
			err = ValidateCPUPeriod(p)
			if err == nil {
				t.Fatalf("expect validate cpu peroid %v error, but err is nil", p)
			}
		}
	}
}

func TestValidateCPUQuota(t *testing.T) {
	// TODO
}
