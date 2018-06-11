package opt

import (
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/stretchr/testify/assert"
)

func TestNewRuntime(t *testing.T) {
	assert := assert.New(t)

	for _, r := range []*map[string]types.Runtime{
		nil,
		{},
		{
			"a": {},
			"b": {Path: "foo"},
		},
	} {
		runtime := NewRuntime(r)
		// just test no panic here
		assert.NoError(runtime.Set("foo=bar"))
	}
}
