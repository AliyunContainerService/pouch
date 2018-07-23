package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"testing"

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
	// TODO
}

func TestGetUnknownFlags(t *testing.T) {
	assert := assert.New(t)
	var fileFlags map[string]interface{}
	var flagSet *pflag.FlagSet
	assert.Equal(nil, getUnknownFlags(flagSet, fileFlags))

	flagSet = pflag.NewFlagSet("flagSet", pflag.ContinueOnError)
	flagSet.String("a", "a", "")
	flagSet.String("b", "b", "")
	flagSet.String("c", "c", "")
	flagSet.Parse([]string{"--a=a", "--b=b", "--c=c"})
	assert.Equal(nil, getUnknownFlags(flagSet, fileFlags))

	fileFlags = map[string]interface{}{
		"a": "a",
		"b": "b",
		"c": "c",
	}
	assert.Equal(nil, getUnknownFlags(flagSet, fileFlags))

	fileFlags = map[string]interface{}{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
		"e": "e",
	}
	assert.Equal(fmt.Errorf("unknown flags: d, e"), getUnknownFlags(flagSet, fileFlags))
}
