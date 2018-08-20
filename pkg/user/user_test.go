package user

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseString(t *testing.T) {
	assert := assert.New(t)
	for _, l := range []string{
		"root:x:0:0",
		"daemon:x:1:1",
		"admin:x:500:500",
	} {
		var u1, u2 string
		var i1, i2 int
		ParseString(l, &u1, &u2, &i1, &i2)
		assert.Equal(fmt.Sprintf("%s:%s:%d:%d", u1, u2, i1, i2), l)
	}
}

func TestGetAdditionalGids(t *testing.T) {
	groups := []string{"1234", "5678"}
	expected := []uint32{1234, 5678}

	result := GetAdditionalGids(groups)
	assert := assert.New(t)
	assert.True(reflect.DeepEqual(expected, result), true)
}
