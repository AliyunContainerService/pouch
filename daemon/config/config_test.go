package config

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestIterateConfig(t *testing.T) {
	assert := assert.New(t)
	origin := map[string]interface{}{
		"a": "a",
		"b": "b",
		"c": "c",
		"iter1": map[string]interface{}{
			"i1": "i1",
			"i2": "i2",
		},
		"iter11": map[string]interface{}{
			"ii1": map[string]interface{}{
				"iii1": "iii1",
				"iii2": "iii2",
			},
		},
	}

	expect := map[string]interface{}{
		"a":    "a",
		"b":    "b",
		"c":    "c",
		"i1":   "i1",
		"i2":   "i2",
		"iii1": "iii1",
		"iii2": "iii2",
	}

	config := make(map[string]interface{})
	iterateConfig(origin, config)
	assert.Equal(config, expect)

	// test nil map will not cause panic
	config = make(map[string]interface{})
	iterateConfig(nil, config)
	assert.Equal(config, map[string]interface{}{})
}

func TestConfigValidate(t *testing.T) {
	// TODO
}

func TestGetConflictConfigurations(t *testing.T) {
	assert := assert.New(t)
	origin1 := &pflag.FlagSet{}
	origin2 := make(map[string]interface{})
	res := getConflictConfigurations(origin1, origin2)
	assert.Equal(res, nil)

	fileFlags := map[string]interface{}{
		"a": "1",
		"b": "2",
	}
	origin3 := &pflag.FlagSet{}
	origin3.String("a", "1", "")
	origin3.String("b", "2", "")
	origin3.String("c", "3", "")

	assert.Equal(nil, getConflictConfigurations(origin3, fileFlags))

	origin3.Set("a", "2")
	assert.Error(getConflictConfigurations(origin3, fileFlags))
	assert.Equal("found conflict flags in command line and config file: from flag: 2 and from config file: 1",
		getConflictConfigurations(origin3, fileFlags).Error())

	origin3.Set("b", "1")
	assert.Error(getConflictConfigurations(origin3, fileFlags))
	assert.Equal("found conflict flags in command line and config file: from flag: 2 and from config file: 1, from flag: 1 and from config file: 2",
		getConflictConfigurations(origin3, fileFlags).Error())

	origin4 := &pflag.FlagSet{}
	origin4.IntSlice("coordinate", []int{1, 2}, "")
	assert.Equal(nil, getConflictConfigurations(origin4, fileFlags))
}

func TestGetUnknownFlags(t *testing.T) {
	// TODO
}
