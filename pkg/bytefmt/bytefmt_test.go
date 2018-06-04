package bytefmt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestByteSize tests function ByteSize()
func TestByteSize(t *testing.T) {
	tests := []struct {
		input  uint64
		expect string
	}{
		{
			input:  0,
			expect: "0",
		},
		{
			input:  1,
			expect: "1B",
		},
		{
			input:  1023,
			expect: "1023B",
		},
		{
			input:  1024,
			expect: "1K",
		},
		{
			input:  1024 * 1023,
			expect: "1023K",
		},
		{
			input:  1024 * 1024,
			expect: "1M",
		},
		{
			input:  1024 * 1024 * 1023,
			expect: "1023M",
		},
		{
			input:  1024 * 1024 * 1024,
			expect: "1G",
		},
		{
			input:  1024 * 1024 * 1024 * 1023,
			expect: "1023G",
		},
		{
			input:  1024 * 1024 * 1024 * 1024,
			expect: "1T",
		},
		{
			input:  1024 * 1024 * 1024 * 1024 * 1500,
			expect: "1500T",
		},
	}

	for _, test := range tests {
		out := ByteSize(test.input)
		assert.Equal(t, test.expect, out)
	}
}

// TestToBytes test TestToBytes.
func TestToBytes(t *testing.T) {
	tests := []struct {
		input  string
		expect uint64
		err    error
	}{
		{
			input:  "",
			expect: 0,
			err:    ErrorInvalidByte,
		},
		{
			input:  "-1",
			expect: 0,
			err:    ErrorInvalidByte,
		},
		{
			input:  "1B",
			expect: 1,
			err:    nil,
		},
		{
			input:  "1k",
			expect: 1024,
			err:    nil,
		},
		{
			input:  "10.5k",
			expect: uint64(10.5 * 1024),
			err:    nil,
		},
		{
			input:  "1024000",
			expect: 1024000,
			err:    nil,
		},
	}
	for _, test := range tests {
		out, err := ToBytes(test.input)
		assert.Equal(t, test.expect, out)
		assert.Equal(t, test.err, err)
	}
}

// TestToMegabytes tests TestToMegabytes.
func TestToMegabytes(t *testing.T) {
	tests := []struct {
		input  string
		expect uint64
		err    error
	}{
		{
			input:  "",
			expect: 0,
			err:    ErrorInvalidByte,
		},
		{
			input:  "-1",
			expect: 0,
			err:    ErrorInvalidByte,
		},
		{
			input:  "10.9m",
			expect: 10,
			err:    nil,
		},
		{
			input:  "1024k",
			expect: 1,
			err:    nil,
		},
	}
	for _, test := range tests {
		out, err := ToMegabytes(test.input)
		assert.Equal(t, test.expect, out)
		assert.Equal(t, test.err, err)
	}
}

// TestToKilobytes tests TestToKilobytes.
func TestToKilobytes(t *testing.T) {
	tests := []struct {
		input  string
		expect uint64
		err    error
	}{
		{
			input:  "",
			expect: 0,
			err:    ErrorInvalidByte,
		},
		{
			input:  "-1",
			expect: 0,
			err:    ErrorInvalidByte,
		},
		{
			input:  "10.5k",
			expect: 10,
			err:    nil,
		},
		{
			input:  "1024B",
			expect: 1,
			err:    nil,
		},
	}
	for _, test := range tests {
		out, err := ToKilobytes(test.input)
		assert.Equal(t, test.expect, out)
		assert.Equal(t, test.err, err)
	}
}
