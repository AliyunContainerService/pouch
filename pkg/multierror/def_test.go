package multierror

import (
	"fmt"
	"testing"
)

func TestMultiErrors(t *testing.T) {
	type tCase struct {
		name string
		errs []error

		expectedSize  int
		expectedError string
	}

	for _, tc := range []tCase{
		{
			name:          "no error",
			errs:          nil,
			expectedSize:  0,
			expectedError: "no error",
		}, {
			name:          "one error",
			errs:          []error{fmt.Errorf("oops1")},
			expectedSize:  1,
			expectedError: "oops1",
		}, {
			name:          "two errors",
			errs:          []error{fmt.Errorf("oops1"), fmt.Errorf("oops2")},
			expectedSize:  2,
			expectedError: "2 errors:\n\n* oops1\n* oops2",
		},
	} {
		merrs := new(Multierrors)
		merrs.Append(tc.errs...)

		if got := merrs.Size(); got != tc.expectedSize {
			t.Fatalf("%s: want %d, but got %d", tc.name, tc.expectedSize, got)
		}

		if got := merrs.Error(); got != tc.expectedError {
			t.Fatalf("%s: want %s, but got %s", tc.name, tc.expectedError, got)
		}
	}
}
