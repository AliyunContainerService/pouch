package logger

import (
	"reflect"
	"strings"
	"testing"
)

func TestInfoExtractAttributes(t *testing.T) {
	info := Info{
		LogConfig: map[string]string{
			"env":    "app",
			"labels": "VERSION",
		},
		ContainerEnvs:   []string{"app=pouchcontainer"},
		ContainerLabels: map[string]string{"VERSION": "unknown-oops"},
	}

	for idx, tc := range []struct {
		keyMod   func(string) string
		expected map[string]string
	}{
		{
			keyMod: nil,
			expected: map[string]string{
				"app":     "pouchcontainer",
				"VERSION": "unknown-oops",
			},
		}, {
			keyMod: strings.Title,
			expected: map[string]string{
				"App":     "pouchcontainer",
				"VERSION": "unknown-oops",
			},
		}, {
			keyMod: strings.ToUpper,
			expected: map[string]string{
				"APP":     "pouchcontainer",
				"VERSION": "unknown-oops",
			},
		},
	} {
		got, err := info.ExtraAttributes(tc.keyMod)
		if err != nil {
			t.Fatalf("[%d case] expect no error here, but got error: %v", idx, err)
		}

		if !reflect.DeepEqual(got, tc.expected) {
			t.Fatalf("[%d case] expect %v, but got %v", idx, tc.expected, got)
		}
	}
}
