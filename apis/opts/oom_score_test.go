package opts

import (
	"testing"
)

func TestValidateOOMScore(t *testing.T) {

	errLess := ValidateOOMScore(-1100)
	testBetween := ValidateOOMScore(500)
	errMore := ValidateOOMScore(1100)

	if errLess == nil || errMore == nil {
		t.Fatal("expect get driver not found error, but err is nil")
	}
	if testBetween != nil {
		t.Fatal("validate OOM score error")
	}

}
