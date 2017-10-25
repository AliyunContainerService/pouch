package serializer

import (
	"io"
)

// Serializer is an interface generalizes object serialize operation.
type Serializer interface {
	Encode(obj Object) ([]byte, error)
	Decode(data []byte, obj Object) error
	EncodeToStream(w io.Writer, obj Object) error
	DecodeFromStream(r io.Reader, obj Object) error
}

// Object is an interface to define object.
type Object interface{}

// Codec is an instance of Serializer.
var Codec = NewSerializer()
