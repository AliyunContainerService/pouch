package kernel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetKernelVersion(t *testing.T) {
	version, err := GetKernelVersion()
	assert.Equal(t, nil, err)

	println(version.String())
}
