package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type tCase struct {
	name     string
	input    int64
	expected string
	err      error
}

func TestFormatTimeInterval(t *testing.T) {

	for _, tc := range []tCase{
		{
			name:     "second",
			input:    time.Now().Add(0 - Second).UnixNano(),
			expected: "1 second",
			err:      nil,
		}, {
			name:     "minute",
			input:    time.Now().Add(0 - Minute).UnixNano(),
			expected: "1 minute",
			err:      nil,
		}, {
			name:     "hour",
			input:    time.Now().Add(0 - Hour).UnixNano(),
			expected: "1 hour",
			err:      nil,
		}, {
			name:     "day",
			input:    time.Now().Add(0 - Day).UnixNano(),
			expected: "1 day",
			err:      nil,
		}, {
			name:     "week",
			input:    time.Now().Add(0 - Week).UnixNano(),
			expected: "1 week",
			err:      nil,
		}, {
			name:     "month",
			input:    time.Now().Add(0 - Month).UnixNano(),
			expected: "1 month",
			err:      nil,
		}, {
			name:     "year",
			input:    time.Now().Add(0 - Year).UnixNano(),
			expected: "1 year",
			err:      nil,
		},
		{
			name:     "seconds",
			input:    time.Now().Add(0 - Second*3).UnixNano(),
			expected: "3 seconds",
			err:      nil,
		}, {
			name:     "minutes",
			input:    time.Now().Add(0 - Minute*3).UnixNano(),
			expected: "3 minutes",
			err:      nil,
		}, {
			name:     "hours",
			input:    time.Now().Add(0 - Hour*3).UnixNano(),
			expected: "3 hours",
			err:      nil,
		}, {
			name:     "days",
			input:    time.Now().Add(0 - Day*3).UnixNano(),
			expected: "3 days",
			err:      nil,
		}, {
			name:     "weeks",
			input:    time.Now().Add(0 - Week*3).UnixNano(),
			expected: "3 weeks",
			err:      nil,
		}, {
			name:     "months",
			input:    time.Now().Add(0 - Month*3).UnixNano(),
			expected: "3 months",
			err:      nil,
		}, {
			name:     "years",
			input:    time.Now().Add(0 - Year*3).UnixNano(),
			expected: "3 years",
			err:      nil,
		}, {
			name:     "invalid",
			input:    time.Now().Add(Second).UnixNano(),
			expected: "",
			err:      errInvalid,
		},
	} {
		output, err := FormatTimeInterval(0, tc.input)
		assert.Equal(t, tc.err, err, tc.name)
		assert.Equal(t, tc.expected, output, tc.name)
	}
}

func TestGetTimestamp(t *testing.T) {
	now := time.Now().In(time.UTC)

	tCases := []struct {
		val      string
		expected string
		hasError bool
	}{
		// relative time
		{"1s", fmt.Sprintf("%d", now.Add(-1*time.Second).Unix()), false},
		{"1m", fmt.Sprintf("%d", now.Add(-1*time.Minute).Unix()), false},
		{"1.5h", fmt.Sprintf("%d", now.Add(-90*time.Minute).Unix()), false},
		{"1h30m", fmt.Sprintf("%d", now.Add(-90*time.Minute).Unix()), false},

		// time
		{"2018-07-16T08:00:00.999999999+08:00", "1531699200.999999999", false},
		{"2018-07-16T08:00:00.999999999+00:00", "1531728000.999999999", false},
		{"2018-07-16T08:00:00.999999999-00:00", "1531728000.999999999", false},
		{"2018-07-16T08:00:00.999999999Z", "1531728000.999999999", false},
		{"2018-07-16T08:00:00.999999999", "1531728000.999999999", false},

		{"2018-07-16T08:00:00", "1531728000.000000000", false},
		{"2018-07-16T08:00:00Z", "1531728000.000000000", false},
		{"2018-07-16T08:00:00+00:00", "1531728000.000000000", false},
		{"2018-07-16T08:00:00-00:00", "1531728000.000000000", false},

		{"2018-07-16T08:00", "1531728000.000000000", false},
		{"2018-07-16T08:00Z", "1531728000.000000000", false},
		{"2018-07-16T08:00+00:00", "1531728000.000000000", false},
		{"2018-07-16T08:00-00:00", "1531728000.000000000", false},

		{"2018-07-16T08", "1531728000.000000000", false},
		{"2018-07-16T08Z", "1531728000.000000000", false},
		{"2018-07-16T08+01:00", "1531724400.000000000", false},
		{"2018-07-16T08-01:00", "1531731600.000000000", false},

		{"2018-07-16", "1531699200.000000000", false},
		{"2018-07-16Z", "1531699200.000000000", false},
		{"2018-07-16+01:00", "1531695600.000000000", false},
		{"2018-07-16-01:00", "1531702800.000000000", false},

		// timestamp
		{"0", "0", false},
		{"12", "12", false},
		{"12a", "12a", false},

		// invalid input
		{"-12", "", true},
		{"2006-01-02T15:04:0Z", "", true},
		{"2006-01-02T15:04:0", "", true},
		{"2006-01-02T15:0Z", "", true},
		{"2006-01-02T15:0", "", true},
	}

	for _, tc := range tCases {
		got, err := GetUnixTimestamp(tc.val, now)
		if err != nil && !tc.hasError {
			t.Fatalf("unexpected error %v", err)
		}

		if err == nil && tc.hasError {
			t.Fatal("expected error, but got nothing")
		}

		if got != tc.expected {
			t.Errorf("expected %v, but got %v", tc.expected, got)
		}
	}
}

func TestParseTimestamp(t *testing.T) {
	tCases := []struct {
		val          string
		defaultSec   int64
		expectedSec  int64
		expectedNano int64
		hasError     bool
	}{
		{"20180510", 0, 20180510, 0, false},
		{"20180510.000000001", 0, 20180510, 1, false},
		{"20180510.0000000010", 0, 20180510, 1, false},
		{"20180510.00000001", 0, 20180510, 10, false},
		{"foo.bar", 0, 0, 0, true},
		{"20180510.bar", 0, 0, 0, true},
		{"", -1, -1, 0, false},
	}

	for _, tc := range tCases {
		s, n, err := ParseTimestamp(tc.val, tc.defaultSec)
		if err == nil && tc.hasError {
			t.Fatal("expected error, but got nothing")
		}

		if err != nil && !tc.hasError {
			t.Fatalf("unexpected error %v", err)
		}

		if s != tc.expectedSec || n != tc.expectedNano {
			t.Fatalf("expected sec %v, nano %v, but got sec %v, nano %v", tc.expectedSec, tc.expectedNano, s, n)
		}
	}
}
