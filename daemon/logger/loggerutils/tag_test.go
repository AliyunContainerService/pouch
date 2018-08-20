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
		ContainerEnvs: []string{
			"APP=pouchcontainer", "VERSION=unknown-oops",
		},
		ContainerLabels: map[string]string{
			"from": "open-source",
			"to":   "all",
		},
		DaemonName: "pouch daemon",
	}

	defaultTemplate := "{{.ID}}"
	for idx, tc := range []struct {
		tag      string
		labels   string
		env      string
		envRegex string
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

		{
			tag:      "{{with .ExtraAttributes nil}}this line should not show up{{end}}",
			expected: "",
			hasError: false,
		},

		// use labels config
		{
			tag:      "{{with .ExtraAttributes nil}}label from is {{.from}}{{end}}",
			labels:   "from,to",
			expected: "label from is open-source",
			hasError: false,
		},

		// use env or env-regex config
		{
			tag:      "{{with .ExtraAttributes nil}}APP is {{.APP}}{{end}}",
			env:      "APP",
			expected: "APP is pouchcontainer",
			hasError: false,
		}, {
			tag:      "{{with .ExtraAttributes nil}}APP is {{.APP}}{{end}}",
			env:      "VERSION",
			expected: "APP is <no value>",
			hasError: false,
		}, {
			tag:      "{{with .ExtraAttributes nil}}Version is {{.VERSION}}{{end}}",
			envRegex: "^[A-Z]+$",
			expected: "Version is unknown-oops",
			hasError: false,
		},
	} {
		info.LogConfig["tag"] = tc.tag
		info.LogConfig["labels"] = tc.labels
		info.LogConfig["env"] = tc.env
		info.LogConfig["env-regex"] = tc.envRegex

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
