package opts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSysctls(t *testing.T) {
	type result struct {
		sysctls map[string]string
		err     error
	}
	type TestCases struct {
		input  []string
		expect result
	}

	testCases := []TestCases{
		{
			input: []string{"a=b"},
			expect: result{
				sysctls: map[string]string{"a": "b"},
				err:     nil,
			},
		},
		{
			input: []string{"ab"},
			expect: result{
				sysctls: nil,
				err:     fmt.Errorf("invalid sysctl %s: sysctl must be in format of key=value", "ab"),
			},
		},
	}

	for _, testCase := range testCases {
		sysctl, err := ParseSysctls(testCase.input)
		assert.Equal(t, testCase.expect.sysctls, sysctl)
		assert.Equal(t, testCase.expect.err, err)
	}
}
