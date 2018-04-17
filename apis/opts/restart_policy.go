package opts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
)

// ParseRestartPolicy parses restart policy param of container.
func ParseRestartPolicy(restartPolicy string) (*types.RestartPolicy, error) {
	policy := &types.RestartPolicy{}

	if restartPolicy == "" {
		policy.Name = "no"
		return policy, nil
	}

	fields := strings.Split(restartPolicy, ":")
	policy.Name = fields[0]

	switch policy.Name {
	case "always", "unless-stopped", "no":
	case "on-failure":
		if len(fields) > 2 {
			return nil, fmt.Errorf("invalid restart policy: %s", restartPolicy)
		}
		if len(fields) == 2 {
			n, err := strconv.Atoi(fields[1])
			if err != nil {
				return nil, fmt.Errorf("invalid restart policy: %v", err)
			}
			policy.MaximumRetryCount = int64(n)
		}
	default:
		return nil, fmt.Errorf("invalid restart policy: %s", restartPolicy)
	}

	return policy, nil
}

// ValidateRestartPolicy verifies the correctness of restart policy of container.
func ValidateRestartPolicy(policy *types.RestartPolicy) error {
	switch policy.Name {
	case "always", "unless-stopped", "no":
		if policy.MaximumRetryCount != 0 {
			return fmt.Errorf("maximum retry count can not be used with restart policy: %s", policy.Name)
		}
	case "on-failure":
		if policy.MaximumRetryCount < 0 {
			return fmt.Errorf("maximum retry count can not be negative")
		}
	case "":
		// ignore
	default:
		return fmt.Errorf("invalid restart policy '%s'", policy.Name)
	}

	return nil
}
