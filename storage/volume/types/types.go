package types

// Option represents volume option struct.
type Option struct {
	Value string
	Desc  string
}

var (
	// OptionRef defines the reference of containers.
	OptionRef = "ref"

	// DefaultBackend defines the default volume backend.
	DefaultBackend = "local"
)
