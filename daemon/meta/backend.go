package meta

var backend Backend

// Backend is an interface which describes what a store should support.
type Backend interface {
	// meta store contructor.
	New(Config) error

	// Put write key-value into store.
	Put(bucket string, key string, value []byte) error

	// Get read object from store.
	Get(bucket string, key string) ([]byte, error)

	// Remove remove all data of the key.
	Remove(key string) error

	// List return all objects with specify bucket.
	List(bucket string) ([][]byte, error)

	// Keys return all keys.
	Keys() ([]string, error)
}

// Register registers a backend to be daemon's store.
func Register(b Backend) {
	backend = b
}
