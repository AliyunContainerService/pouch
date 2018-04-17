package opts

import "fmt"

// ValidateOOMScore validates oom score
func ValidateOOMScore(score int64) error {
	if score < -1000 || score > 1000 {
		return fmt.Errorf("oom-score-adj should be in range [-1000, 1000]")
	}

	return nil
}
