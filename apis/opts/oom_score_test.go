package opts

import "testing"

func TestValidateOOMScore(t *testing.T) {
      cases := []struct {
          in, want int64
      }{
          {-1001, 0},
	  {500,0},
          {1001,0}, 
      }
      for _, c := range cases {
          got := ValidateCPUPeriod(c.in)
          if got != nil {
              t.Errorf("oom-score-adj should be in range [-1000, 1000]")
          }
}

