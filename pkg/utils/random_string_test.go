package utils

import (
	"testing"
	"time"

	"github.com/alibaba/pouch/pkg/log"
)

func TestRandString(t *testing.T) {
	start := time.Now()
	results := []string{}
	for i := 0; i < 1000; i++ {
		str := RandString(8, "", "")
		if StringInSlice(results, str) {
			t.Errorf("RandString got a same random string in the test: %s", str)
		}

		results = append(results, str)
	}
	end := time.Now()
	elapsed := end.Sub(start)
	log.With(nil).Infof("TestRandString generate 1000 random strings costs: %v s", elapsed.Seconds())
}
