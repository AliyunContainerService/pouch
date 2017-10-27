package jsonstream

import (
	"bytes"
	"encoding/json"
)

// Formater define the data's format via stream transfer.
type Formater interface {
	// BegineWrite write some datas at the stream beginning.
	BeginWrite() ([]byte, error)

	// EndWrite write some datas at the stream ending.
	EndWrite() ([]byte, error)

	// Write write a object be serialized.
	Write(o interface{}) ([]byte, error)
}

var (
	jsonBeginDelim = []byte("[")
	jsonEndDelim   = []byte("]")
	jsonSep        = []byte(",")
)

// defaultFormat define the pouch's pull progress, is not compatible with docker's pull.
type defaultFormat struct {
	beginDelim []byte
	endDelim   []byte
	sep        []byte
	isFirst    bool
}

func newDefaultFormat() Formater {
	return &defaultFormat{
		beginDelim: jsonBeginDelim,
		endDelim:   jsonEndDelim,
		sep:        jsonSep,
		isFirst:    true,
	}
}

func (f *defaultFormat) BeginWrite() ([]byte, error) {
	return f.beginDelim, nil
}

func (f *defaultFormat) EndWrite() ([]byte, error) {
	return f.endDelim, nil
}

func (f *defaultFormat) Write(o interface{}) ([]byte, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	if !f.isFirst {
		buf := bytes.NewBuffer(f.sep)
		if _, err := buf.Write(b); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	f.isFirst = false

	return b, nil
}
