package randomid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	id := Generate()
	assert.Equal(t, 64, len(id))
}
