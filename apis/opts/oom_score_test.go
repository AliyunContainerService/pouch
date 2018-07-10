package opts

import (
	"testing"
	"math/rand"
	"time"
)

func TestValidateOOMScoreSucceed(t *testing.T) {
	num1 := int(2000)
	num2 := int(1000)

	var i int
	for i = 0; i < 5; i++ {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		e := ValidateOOMScore(int64(r.Intn(num1) - num2))
		if e != nil {
			t.Error("test failed")
			return
		}
	}
}

func TestValidateOOMScoreFailed(t *testing.T) {
	uvinf := 0x7FF00000000000
	uvneginf := 0xFFF00000000000
	num1 := int(1000)

	var j int
	for j = 0; j < 5; j++ {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		e := ValidateOOMScore(int64(uvinf - r.Intn(num1)))
		eNew := ValidateOOMScore(- int64(uvneginf - r.Intn(num1)))
		if e != nil || eNew != nil{
			t.Error("test failed")
			return
		}
	}
}
