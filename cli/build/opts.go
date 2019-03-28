package build

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/pkg/reference"
)

// from frontend dockerfile codebase
var (
	keyTarget         = "target"
	keyBuildArgPrefix = "build-arg:"
)

// Options is used to contains the user setting for build.
type Options struct {
	Target    string
	BuildArgs map[string]string
	TagList   []string
	LocalDirs map[string]string
}

// optsToFrontendAttrs converts build options to FrontendAttrs.
func optsToFrontendAttrs(opt *Options) (map[string]string, error) {
	attrs := map[string]string{}

	// target setting
	if opt.Target != "" {
		attrs[keyTarget] = opt.Target
	}

	// add build-args
	for key, value := range opt.BuildArgs {
		attrs[keyBuildArgPrefix+key] = value
	}
	return attrs, nil
}

// optsToExporterAttrs converts build options tp ExporterAttrs.
func optsToExporterAttrs(opt *Options) (map[string]string, error) {
	attrs := map[string]string{}

	// apply tag list
	if len(opt.TagList) == 0 {
		return nil, fmt.Errorf("missing exporter name")
	}
	tagList := make([]string, 0, len(opt.TagList))
	for _, tag := range opt.TagList {
		namedRef, err := reference.Parse(tag)
		if err != nil {
			return nil, fmt.Errorf("failed to parse reference %s: %v", tag, err)
		}

		namedRef = reference.TrimTagForDigest(reference.WithDefaultTagIfMissing(namedRef))
		tagList = append(tagList, namedRef.String())
	}
	attrs["name"] = strings.Join(tagList, ",")
	return attrs, nil
}
