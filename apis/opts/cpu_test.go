package opts

import (
  "fmt"
  "testing"

  "github.com/stretchr/testify/assert"
)

func TestValidateCPUQuota(t *testing.T) {
  type TestCase struct {
    quota    int64
    expected error
  }

  testCases := []TestCase{
    {
      quota:    0,
      expected: nil,
    },
    {
      quota:    -1,
      expected: fmt.Errorf("CPU cfs quota %d cannot be less than 1ms (i.e. 1000)", -1),
    },
    {
      quota:    1001,
      expected: nil,
    },
  }

  for _, testCase := range testCases {
    err := ValidateCPUQuota(testCase.quota)
    assert.Equal(t, testCase.expected, err)
  }
}