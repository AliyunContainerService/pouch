package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeBody(t *testing.T) {
	assert := assert.New(t)

	obj := &struct {
		a string
	}{
		a: "bar",
	}

	d, err := json.Marshal(obj)
	assert.NoError(err)
	body := ioutil.NopCloser(bytes.NewReader([]byte(d)))
	assert.NoError(decodeBody(obj, body))
}

func TestEnsureCloseReader(t *testing.T) {
	// test input is nil or not should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ensureCloseReader should not panic")
		}
	}()

	resp := &Response{}

	ensureCloseReader(resp)
	ensureCloseReader(nil)
}
