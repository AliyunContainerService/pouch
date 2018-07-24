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

	flagSet := new(pflag.FlagSet)
	flagSet.String("normal", "", "normal flag")
	flagSet.Set("normal", "normal")

	flagSet.Bool("conflict", false, "conflict flag")
	flagSet.Set("conflict", "true")

	flagSet.String("notSet", "notSet", "flag not set")

	flagSet.IntSlice("intSlice", []int{1, 2, 3}, "skip int slice")
	flagSet.Set("intSlice", "[1,2,3]")

	noConflictFlagMap := map[string]interface{}{
		"a": "a",
		"b": "b",
		"c": "c",
	}

	conflictFlagMap := map[string]interface{}{
		"a":        "a",
		"conflict": false,
	}

	skipUnsetFlagMap := map[string]interface{}{
		"a":      "a",
		"notSet": "notSet",
	}

	skipSliceFlagMap := map[string]interface{}{
		"a":        100,
		"intSlice": []int{1, 2, 3},
	}

	error := getConflictConfigurations(flagSet, noConflictFlagMap)
	assert.Equal(error, nil)

	error = getConflictConfigurations(flagSet, conflictFlagMap)
	assert.NotNil(error)

	error = getConflictConfigurations(flagSet, skipUnsetFlagMap)
	assert.Equal(error, nil)

	error = getConflictConfigurations(flagSet, skipSliceFlagMap)
	assert.Equal(error, nil)
}

func TestGetUnknownFlags(t *testing.T) {
	// TODO
}
