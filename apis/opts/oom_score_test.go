package opts

import (
  "fmt"
  "testing"

  "github.com/stretchr/testify/assert"
)

func TestValidateOOMScore(t *testing.T) {
  // TODO
  tests := []struct {
    score   int64
    res    error
  }{
    // TODO: Add test cases.
    {-1001, fmt.Errorf("oom-score-adj should be in range [-1000, 1000]")},
    {0, nil},
    {1001, fmt.Errorf("oom-score-adj should be in range [-1000, 1000]")},
  }
  for _, tt := range tests {
    res := ValidateOOMScore(tt.score)
    assert.Equal(t, res, tt.res)
  }
}