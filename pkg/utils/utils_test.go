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

func TestFormatCreatedTime(t *testing.T) {

	for _, tc := range []tCase{
		{
			name:     "second",
			input:    time.Now().Add(0 - Second).UTC().Format(TimeLayout),
			expected: "about 1 second ago",
			err:      nil,
		}, {
			name:     "minute",
			input:    time.Now().Add(0 - Minute).UTC().Format(TimeLayout),
			expected: "about 1 minute ago",
			err:      nil,
		}, {
			name:     "hour",
			input:    time.Now().Add(0 - Hour).UTC().Format(TimeLayout),
			expected: "about 1 hour ago",
			err:      nil,
		}, {
			name:     "day",
			input:    time.Now().Add(0 - Day).UTC().Format(TimeLayout),
			expected: "about 1 day ago",
			err:      nil,
		}, {
			name:     "week",
			input:    time.Now().Add(0 - Week).UTC().Format(TimeLayout),
			expected: "about 1 week ago",
			err:      nil,
		}, {
			name:     "month",
			input:    time.Now().Add(0 - Month).UTC().Format(TimeLayout),
			expected: "about 1 month ago",
			err:      nil,
		}, {
			name:     "year",
			input:    time.Now().Add(0 - Year).UTC().Format(TimeLayout),
			expected: "about 1 year ago",
			err:      nil,
		},
		{
			name:     "seconds",
			input:    time.Now().Add(0 - Second*3).UTC().Format(TimeLayout),
			expected: "about 3 seconds ago",
			err:      nil,
		}, {
			name:     "minutes",
			input:    time.Now().Add(0 - Minute*3).UTC().Format(TimeLayout),
			expected: "about 3 minutes ago",
			err:      nil,
		}, {
			name:     "hours",
			input:    time.Now().Add(0 - Hour*3).UTC().Format(TimeLayout),
			expected: "about 3 hours ago",
			err:      nil,
		}, {
			name:     "days",
			input:    time.Now().Add(0 - Day*3).UTC().Format(TimeLayout),
			expected: "about 3 days ago",
			err:      nil,
		}, {
			name:     "weeks",
			input:    time.Now().Add(0 - Week*3).UTC().Format(TimeLayout),
			expected: "about 3 weeks ago",
			err:      nil,
		}, {
			name:     "months",
			input:    time.Now().Add(0 - Month*3).UTC().Format(TimeLayout),
			expected: "about 3 months ago",
			err:      nil,
		}, {
			name:     "years",
			input:    time.Now().Add(0 - Year*3).UTC().Format(TimeLayout),
			expected: "about 3 years ago",
			err:      nil,
		}, {
			name:     "notHappen",
			input:    time.Now().Add(Second * 61).UTC().Format(TimeLayout),
			expected: "about -1 minute ago",
			err:      nil,
		},
		{
			name:     "invalid",
			input:    "balabala",
			expected: "",
			err:      errInvalid,
		},
	} {
		output, err := FormatCreatedTime(tc.input)
		assert.Equal(t, tc.err, err, tc.name)
		assert.Equal(t, tc.expected, output, tc.name)
	}

}
