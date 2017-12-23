package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type tCase struct {
	name     string
	input    string
	expected string
	err      error
}

func TestFormatSize(t *testing.T) {
	assert := assert.New(t)
	kvs := map[int64]string{
		-1024:         "0.00 B",
		0:             "0.00 B",
		1024:          "1.00 KB",
		1024000:       "1000.00 KB",
		1024000000000: "953.67 GB",
	}

	for k, v := range kvs {
		size := FormatSize(k)
		assert.Equal(v, size)
	}
}

func TestFormatTimeInterval(t *testing.T) {

	for _, tc := range []tCase{
		{
			name:     "second",
			input:    time.Now().Add(0 - Second).UTC().Format(TimeLayout),
			expected: "1 second",
			err:      nil,
		}, {
			name:     "minute",
			input:    time.Now().Add(0 - Minute).UTC().Format(TimeLayout),
			expected: "1 minute",
			err:      nil,
		}, {
			name:     "hour",
			input:    time.Now().Add(0 - Hour).UTC().Format(TimeLayout),
			expected: "1 hour",
			err:      nil,
		}, {
			name:     "day",
			input:    time.Now().Add(0 - Day).UTC().Format(TimeLayout),
			expected: "1 day",
			err:      nil,
		}, {
			name:     "week",
			input:    time.Now().Add(0 - Week).UTC().Format(TimeLayout),
			expected: "1 week",
			err:      nil,
		}, {
			name:     "month",
			input:    time.Now().Add(0 - Month).UTC().Format(TimeLayout),
			expected: "1 month",
			err:      nil,
		}, {
			name:     "year",
			input:    time.Now().Add(0 - Year).UTC().Format(TimeLayout),
			expected: "1 year",
			err:      nil,
		},
		{
			name:     "seconds",
			input:    time.Now().Add(0 - Second*3).UTC().Format(TimeLayout),
			expected: "3 seconds",
			err:      nil,
		}, {
			name:     "minutes",
			input:    time.Now().Add(0 - Minute*3).UTC().Format(TimeLayout),
			expected: "3 minutes",
			err:      nil,
		}, {
			name:     "hours",
			input:    time.Now().Add(0 - Hour*3).UTC().Format(TimeLayout),
			expected: "3 hours",
			err:      nil,
		}, {
			name:     "days",
			input:    time.Now().Add(0 - Day*3).UTC().Format(TimeLayout),
			expected: "3 days",
			err:      nil,
		}, {
			name:     "weeks",
			input:    time.Now().Add(0 - Week*3).UTC().Format(TimeLayout),
			expected: "3 weeks",
			err:      nil,
		}, {
			name:     "months",
			input:    time.Now().Add(0 - Month*3).UTC().Format(TimeLayout),
			expected: "3 months",
			err:      nil,
		}, {
			name:     "years",
			input:    time.Now().Add(0 - Year*3).UTC().Format(TimeLayout),
			expected: "3 years",
			err:      nil,
		}, {
			name:     "notHappen",
			input:    time.Now().Add(Second * 61).UTC().Format(TimeLayout),
			expected: "-1 minute",
			err:      nil,
		},
		{
			name:     "invalid",
			input:    "balabala",
			expected: "",
			err:      errInvalid,
		},
	} {
		output, err := FormatTimeInterval(tc.input)
		assert.Equal(t, tc.err, err, tc.name)
		assert.Equal(t, tc.expected, output, tc.name)
	}

}
