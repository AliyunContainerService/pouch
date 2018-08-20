package loggerutils

import (
	"bytes"

	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/pkg/utils/templates"
)

// GenerateLogTag returns a tag which can be used for different log drivers
// based on the container.
func GenerateLogTag(info logger.Info, defaultTemplate string) (string, error) {
	tagTemplate := info.LogConfig["tag"]
	if tagTemplate == "" {
		tagTemplate = defaultTemplate
	}

	tmpl, err := templates.NewParse("logtag", tagTemplate)
	if err != nil {
		return "", err
	}

	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, &info); err != nil {
		return "", err
	}
	return buf.String(), nil
}
