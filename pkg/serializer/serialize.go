package serializer

import (
	"encoding/json"
	"io"
)

// ContentType is a defined type.
type ContentType string

const (
	// ContentTypeJSON is a json ContentType.
	ContentTypeJSON ContentType = "application/json"
	defaultType                 = "application/json"
)

// String return ContentType as string type.
func (c ContentType) String() string {
	return string(c)
}

// Serialization is the default implement of interface Serializer.
type Serialization struct {
}

// NewSerializer returns Serializer instance.
func NewSerializer() Serializer {
	return &Serialization{}
}

func (s *Serialization) encode(obj Object) ([]byte, error) {
	return json.Marshal(obj)
}

func (s *Serialization) decode(data []byte, obj Object) error {
	return json.Unmarshal(data, obj)

}

// Encode is used to encode obj.
func (s *Serialization) Encode(obj Object) ([]byte, error) {
	return s.encode(obj)
}

// Decode is used to decode obj from date.
func (s *Serialization) Decode(data []byte, obj Object) error {
	return s.decode(data, obj)
}

// EncodeToStream is used to encode obj to stream.
func (s *Serialization) EncodeToStream(w io.Writer, obj Object) error {
	b, err := s.Encode(obj)
	if err != nil {
		return err
	}
	w.Write(b)
	w.Write([]byte("\n"))
	return nil
}

// DecodeFromStream is used to decode obj from stream.
func (s *Serialization) DecodeFromStream(r io.Reader, obj Object) error {
	return json.NewDecoder(r).Decode(obj)
}
