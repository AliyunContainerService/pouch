package opts

import (
	"errors"
	"fmt"
	"strings"
)

// ParseLogOptions parses [key=value] slice-type log options into map.
func ParseLogOptions(driver string, logOpts []string) (map[string]string, error) {
	opts, err := convertKVStringsToMap(logOpts)
	if err != nil {
		return nil, err
	}

	if driver == "none" && len(opts) > 0 {
		return nil, fmt.Errorf("don't allow to set logging opts for driver %s", driver)
	}
	return opts, nil
}

// convertKVStringsToMap converts ["key=value"] into {"key":"value"}
//
// TODO(fuwei): make it common in the opts.ParseXXX().
func convertKVStringsToMap(values []string) (map[string]string, error) {
	kvs := make(map[string]string, len(values))

	for _, value := range values {
		terms := strings.SplitN(value, "=", 2)
		if len(terms) != 2 {
			return nil, errors.New("the format must be key=value")
		}
		kvs[terms[0]] = terms[1]
	}
	return kvs, nil
}
