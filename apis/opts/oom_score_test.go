package opts

import (
	"testing"
)

func TestValidateOOMScore(t *testing.T) {

	err_less := ValidateOOMScore(-1100)
	test_between := ValidateOOMScore(500)
	err_more := ValidateOOMScore(1100)

	if err_less == nil || err_more == nil {
		t.Fatal("expect get driver not found error, but err is nil")
	}
	if test_between != nil {
		t.Fatal("validate OOM score error")
	}

}
