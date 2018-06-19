package opts

import (
	"reflect"
	"testing"
)

func TestParseLogOptions(t *testing.T) {
	type tCases struct {
		driver  string
		logOpts []string

		hasError bool
		expected map[string]string
	}

	for idx, tc := range []tCases{
		{
			driver:   "",
			logOpts:  nil,
			hasError: false,
			expected: map[string]string{},
		}, {
			driver:   "none",
			logOpts:  nil,
			hasError: false,
			expected: map[string]string{},
		}, {
			driver:   "none",
			logOpts:  []string{"haha"},
			hasError: true,
			expected: nil,
		}, {
			driver:   "none",
			logOpts:  []string{"test=1"},
			hasError: true,
			expected: nil,
		}, {
			driver:   "json-file",
			logOpts:  []string{"test=1"},
			hasError: false,
			expected: map[string]string{
				"test": "1",
			},
		}, {
			driver:   "json-file",
			logOpts:  []string{"test=1=1"},
			hasError: false,
			expected: map[string]string{
				"test": "1=1",
			},
		}, {
			driver:   "json-file",
			logOpts:  []string{"test=1", "flag=oops", "test=2"},
			hasError: false,
			expected: map[string]string{
				"test": "2",
				"flag": "oops",
			},
		},
	} {
		got, err := ParseLogOptions(tc.driver, tc.logOpts)
		if err == nil && tc.hasError {
			t.Fatalf("[%d case] should have error here, but got nothing", idx)
		}
		if err != nil && !tc.hasError {
			t.Fatalf("[%d case] should have no error here, but got error(%v)", idx, err)
		}

		if !reflect.DeepEqual(got, tc.expected) {
			t.Fatalf("[%d case] should have (%v), but got (%v)", idx, tc.expected, got)
		}
	}
}
