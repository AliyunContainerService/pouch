package types

// Option represents volume option struct.
type Option struct {
	Value string
	Desc  string
}

var (
	// APIVersion defines control server api version.
	APIVersion = "/api/v1"

	// OptionRef defines the reference of containers.
	OptionRef = "ref"

	// DefaultBackend defines the default volume backend.
	DefaultBackend = "local"
)
