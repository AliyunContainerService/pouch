package meta

var backend Backend

// Backend is an interface which describes what a store should support.
type Backend interface {
	New(Config) error
	Put(string, string, []byte) error
	Get(string, string) ([]byte, error)
	Remove(string) error
	List(string) ([][]byte, error)
	Keys() ([]string, error)
}

// Register registers a backend to be daemon's store.
func Register(b Backend) {
	backend = b
}
