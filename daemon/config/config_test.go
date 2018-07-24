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
	// TODO
}

func TestGetUnknownFlags(t *testing.T) {

	flagSet := new(pflag.FlagSet)
	flagSet.String("a", "a", "a")
	flagSet.Bool("b", false, "b")
	flagSet.Int("c", -500, "c")

	flagSetNil := new(pflag.FlagSet)

	assert := assert.New(t)

	fileFlagsKnown := map[string]interface{}{
		"a": "a",
		"b": true,
	}

	fileFlagsUnknown := map[string]interface{}{
		"c": 100,
		"d": "d",
	}

	fileFlagsNil := map[string]interface{}{}

	error := getUnknownFlags(flagSet, fileFlagsKnown)
	assert.Equal(error, nil)

	error = getUnknownFlags(flagSet, fileFlagsUnknown)
	assert.NotNil(error)

	error = getUnknownFlags(flagSet, fileFlagsNil)
	assert.Equal(error, nil)

	error = getUnknownFlags(flagSetNil, fileFlagsUnknown)
	assert.NotNil(error)

	error = getUnknownFlags(flagSetNil, fileFlagsKnown)
	assert.NotNil(error)

	error = getUnknownFlags(flagSetNil, fileFlagsNil)
	assert.Equal(error, nil)

}
