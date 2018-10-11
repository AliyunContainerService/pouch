package metrics

// Unit represents the type or precision of a metric that is appended to
// the metrics fully qualified name
type Unit string

const (
	nanoseconds Unit = "nanoseconds"
	seconds     Unit = "seconds"
	bytes       Unit = "bytes"
	total       Unit = "total"
)
