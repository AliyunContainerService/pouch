package loggerutils

import (
	"testing"

	"github.com/alibaba/pouch/daemon/logger"
)

func TestGenerateLogTag(t *testing.T) {
	info := logger.Info{
		LogConfig:        map[string]string{},
		ContainerID:      "pouchcontainer-20180610",
		ContainerName:    "created_by_$(whois)",
		ContainerImageID: "8c811b4aec35",
		DaemonName:       "pouch daemon",
	}

	defaultTemplate := "{{.ID}}"
	for idx, tc := range []struct {
		tag      string
		expected string
		hasError bool
	}{
		{
			tag:      "",
			expected: info.ID(),
			hasError: false,
		}, {
			tag:      "{{.FullID}}",
			expected: "pouchcontainer-20180610",
			hasError: false,
		}, {
			tag:      "{{.FullID}} - {{.Name}} - {{.ImageID}}",
			expected: "pouchcontainer-20180610 - created_by_$(whois) - 8c811b4aec35",
			hasError: false,
		}, {
			tag:      "{{.DaemonName}}",
			expected: "pouch daemon",
			hasError: false,
		}, {
			tag:      "{{.NotSupport}}",
			expected: "",
			hasError: true,
		},
	} {
		info.LogConfig["tag"] = tc.tag
		got, err := GenerateLogTag(info, defaultTemplate)

		if err != nil && !tc.hasError {
			t.Fatalf("[%d case] expect no error here, but got error: %v", idx, err)
		}

		if err == nil && tc.hasError {
			t.Fatalf("[%d case] expect error here, but got nothing", idx)
		}

		if got != tc.expected {
			t.Fatalf("[%d case] expect value (%v), but got (%v)", idx, tc.expected, got)
		}
	}
}
