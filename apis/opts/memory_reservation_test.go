package opts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMemoryReservation(t *testing.T) {
	type result struct {
		memoryReservation int64
		err               error
	}
	type TestCase struct {
		input    string
		expected result
	}

	testCases := []TestCase{
		{
			input: "",
			expected: result{
				memoryReservation: 0,
				err:               nil,
			},
		},
		{
			input: "0",
			expected: result{
				memoryReservation: 0,
				err:               nil,
			},
		},
		{
			input: "100m",
			expected: result{
				memoryReservation: 104857600,
				err:               nil,
			},
		},
		{
			input: "10invalid",
			expected: result{
				memoryReservation: -1,
				err:               fmt.Errorf("invalid size: '%s'", "10invalid"),
			},
		},
	}

	for _, testCase := range testCases {
		memoryReservation, err := ParseMemoryReservation(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.memoryReservation, memoryReservation)
	}
}
