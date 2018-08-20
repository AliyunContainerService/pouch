package opts

import (
	"fmt"

	"github.com/alibaba/pouch/pkg/utils"
)

// ParseLogOptions parses [key=value] slice-type log options into map.
func ParseLogOptions(driver string, logOpts []string) (map[string]string, error) {
	opts, err := utils.ConvertKVStringsToMap(logOpts)
	if err != nil {
		return nil, err
	}

	if driver == "none" && len(opts) > 0 {
		return nil, fmt.Errorf("don't allow to set logging opts for driver %s", driver)
	}
	return opts, nil
}
