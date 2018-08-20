package logger

import (
	"regexp"
	"strings"

	"github.com/alibaba/pouch/pkg/utils"
)

// Info provides container information for log driver.
//
// TODO(fuwei): add more fields.
type Info struct {
	LogConfig map[string]string

	ContainerID      string
	ContainerName    string
	ContainerImageID string
	ContainerEnvs    []string
	ContainerLabels  map[string]string
	ContainerRootDir string

	DaemonName string
}

// ID returns the container truncated ID.
func (i *Info) ID() string {
	return utils.TruncateID(i.ContainerID)
}

// FullID returns the container ID.
func (i *Info) FullID() string {
	return i.ContainerID
}

// Name returns the container name.
func (i *Info) Name() string {
	return i.ContainerName
}

// ImageID returns the container's image truncated ID.
func (i *Info) ImageID() string {
	return utils.TruncateID(i.ContainerImageID)
}

// ImageFullID returns the container's image ID.
func (i *Info) ImageFullID() string {
	return i.ContainerImageID
}

// ExtraAttributes returns the user-defined extra attributes (labels, environment
// variables) in key-value format.
//
// NOTE: This can be used by log-driver which support metadata in log file.
func (i *Info) ExtraAttributes(keyMod func(string) string) (map[string]string, error) {
	extra := make(map[string]string)

	// extra specific labels from container labels
	labels, ok := i.LogConfig["labels"]
	if ok && len(labels) > 0 {
		for _, l := range strings.Split(labels, ",") {
			if v, ok := i.ContainerLabels[l]; ok {
				if keyMod != nil {
					l = keyMod(l)
				}
				extra[l] = v
			}
		}
	}

	containerEnvs := make(map[string]string)
	for _, e := range i.ContainerEnvs {
		if kv := strings.SplitN(e, "=", 2); len(kv) == 2 {
			containerEnvs[kv[0]] = kv[1]
		}
	}

	// extra specific envs from container envs
	env, ok := i.LogConfig["env"]
	if ok && len(env) > 0 {
		for _, l := range strings.Split(env, ",") {
			if v, ok := containerEnvs[l]; ok {
				if keyMod != nil {
					l = keyMod(l)
				}
				extra[l] = v
			}
		}
	}

	// extra specific envs from container envs by regex
	envRegex, ok := i.LogConfig["env-regex"]
	if ok && len(envRegex) > 0 {
		re, err := regexp.Compile(envRegex)
		if err != nil {
			return nil, err
		}

		for k, v := range containerEnvs {
			if re.MatchString(k) {
				if keyMod != nil {
					k = keyMod(k)
				}
				extra[k] = v
			}
		}
	}

	return extra, nil
}
