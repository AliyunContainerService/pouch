package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatSize(t *testing.T) {
	assert := assert.New(t)
	kvs := map[int64]string{
		-1024:         "0.00 B",
		0:             "0.00 B",
		1024:          "1.00 KB",
		1024000:       "1000.00 KB",
		1024000000000: "953.67 GB",
	}

	for k, v := range kvs {
		size := FormatSize(k)
		assert.Equal(v, size)
	}
}
