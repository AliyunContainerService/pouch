package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRestartPolicy(t *testing.T) {
	type TestCase struct {
		input         string
		expectedName  string
		expectedCount int64
		err           error
	}

	cases := []TestCase{
		{
			input:         "always",
			expectedName:  "always",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "no",
			expectedName:  "no",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "unless-stopped",
			expectedName:  "unless-stopped",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "on-failure:1",
			expectedName:  "on-failure",
			expectedCount: 1,
			err:           nil,
		},
		{
			input:         "on-failure",
			expectedName:  "on-failure",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "on-failure:1:2",
			expectedName:  "on-failure",
			expectedCount: 0,
			err:           fmt.Errorf("invalid restart policy: %s", "on-failure:1:2"),
		},
	}

	for _, cs := range cases {
		policy, err := parseRestartPolicy(cs.input)
		assert.Equal(t, cs.err, err)
		if err == nil {
			assert.Equal(t, cs.expectedName, policy.Name)
			assert.Equal(t, cs.expectedCount, policy.MaximumRetryCount)
		}
	}
}

func TestParseLabels(t *testing.T) {
	type result struct {
		labels map[string]string
		err    error
	}
	type TestCase struct {
		input    []string
		expected result
	}

	testCases := []TestCase{
		{
			input: []string{"a=b"},
			expected: result{
				labels: map[string]string{
					"a": "b",
				},
				err: nil,
			},
		},
		{
			input: []string{"a=b", "a=b"},
			expected: result{
				labels: map[string]string{
					"a": "b",
				},
				err: nil,
			},
		},
		{
			// FIXME: this case should throw error
			input: []string{"a=b", "a=bb"},
			expected: result{
				labels: map[string]string{
					"a": "bb",
				},
				err: nil,
			},
		},
		{
			input: []string{"ThisIsALableWithoutEqualMark"},
			expected: result{
				labels: nil,
				err:    fmt.Errorf("invalid label: %s", "ThisIsALableWithoutEqualMark"),
			},
		},
	}

	for _, testCase := range testCases {
		labels, err := parseLabels(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.labels, labels)
	}
}

func TestParseDeviceMappings(t *testing.T) {

}

func TestParseMemory(t *testing.T) {
	type result struct {
		memory int64
		err    error
	}
	type TestCase struct {
		input    string
		expected result
	}

	testCases := []TestCase{
		{
			input: "",
			expected: result{
				memory: 0,
				err:    nil,
			},
		},
		{
			input: "0",
			expected: result{
				memory: 0,
				err:    nil,
			},
		},
		{
			input: "100m",
			expected: result{
				memory: 104857600,
				err:    nil,
			},
		},
		{
			input: "10asdfg",
			expected: result{
				memory: 0,
				err:    fmt.Errorf("invalid size: '%s'", "10asdfg"),
			},
		},
	}

	for _, testCase := range testCases {
		memory, err := parseMemory(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.memory, memory)
	}
}

func TestParseMemorySwap(t *testing.T) {
	type result struct {
		memorySwap int64
		err        error
	}
	type TestCase struct {
		input    string
		expected result
	}

	testCases := []TestCase{
		{
			input: "",
			expected: result{
				memorySwap: 0,
				err:        nil,
			},
		},
		{
			input: "-1",
			expected: result{
				memorySwap: -1,
				err:        nil,
			},
		},
		{
			input: "100m",
			expected: result{
				memorySwap: 104857600,
				err:        nil,
			},
		},
		{
			input: "10asdfg",
			expected: result{
				memorySwap: 0,
				err:        fmt.Errorf("invalid size: '%s'", "10asdfg"),
			},
		},
	}

	for _, testCase := range testCases {
		memorySwap, err := parseMemorySwap(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.memorySwap, memorySwap)
	}
}

func TestValidateMemorySwappiness(t *testing.T) {
	type TestCase struct {
		input    int64
		expected error
	}

	testCases := []TestCase{
		{
			input:    -1,
			expected: nil,
		},
		{
			input:    0,
			expected: nil,
		},
		{
			input:    100,
			expected: nil,
		},
		{
			input:    38,
			expected: nil,
		},
		{
			input:    -5,
			expected: fmt.Errorf("invalid memory swappiness: %d (its range is -1 or 0-100)", -5),
		},
		{
			input:    200,
			expected: fmt.Errorf("invalid memory swappiness: %d (its range is -1 or 0-100)", 200),
		},
	}

	for _, testCase := range testCases {
		err := validateMemorySwappiness(testCase.input)
		assert.Equal(t, testCase.expected, err)
	}
}
