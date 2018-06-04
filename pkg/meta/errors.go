package meta

const (
	objNotFound = iota

	bucketNotFound
)

var (
	// ErrObjectNotFound is returned when there is no object found.
	ErrObjectNotFound = Error{objNotFound, "Object not found"}

	// ErrBucketNotFound returns the error that no bucket found.
	ErrBucketNotFound = Error{bucketNotFound, "Bucket not found"}
)

// Error is a type of error used for meta.
type Error struct {
	code int
	msg  string
}

// Error returns the message in MetaError.
func (e Error) Error() string {
	return e.msg
}

// IsNotfound return true if code in MetaError is objNotfound.
func (e Error) IsNotfound() bool {
	return e.code == objNotFound
}
