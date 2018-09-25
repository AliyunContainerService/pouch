package utils

// Move the public utility methods which are used by both v1alpha1
// and v1alpha2 here to reduce the code duplication.

// MatchLabelSelector returns true if labels cover selector.
func MatchLabelSelector(selector, labels map[string]string) bool {
	for k, v := range selector {
		if val, ok := labels[k]; ok {
			if v != val {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
