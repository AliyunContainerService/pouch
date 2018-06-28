package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFilter(t *testing.T) {
	assert := assert.New(t)
	type tCase struct {
		filter   []string
		ok       bool
		errorMsg string
	}

	for _, t := range []tCase{
		{
			filter: []string{"id=a", "name=b"},
			ok:     true,
		},
		{
			filter: []string{"status=running"},
			ok:     true,
		},
		{
			filter:   []string{"foo"},
			ok:       false,
			errorMsg: "Bad format of filter, expected name=value",
		},
		{
			filter:   []string{"foo=bar"},
			ok:       false,
			errorMsg: "Invalid filter",
		},
		{
			filter: []string{"label=a", "label=a=b"},
			ok:     true,
		},
		{
			filter: []string{"label=a!=b", "id=aaa"},
			ok:     true,
		},
	} {
		_, err := Parse(t.filter)
		if t.ok {
			assert.NoError(err)
		} else {
			assert.Contains(err.Error(), t.errorMsg)
		}
	}
}
