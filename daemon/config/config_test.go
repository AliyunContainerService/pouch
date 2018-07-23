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
	assert := assert.New(t)

	fileFlags := map[string]interface{}{
		"flag1": "1",
		"flag2": "2",
	}

	flagSet := pflag.NewFlagSet("FlagConfig", pflag.ContinueOnError)
	flagSet.String("flag1", "1", "flag1")
	flagSet.String("flag2", "2", "flag2")
	flagSet.String("flag3", "3", "flag3")
	flagSet.IntSlice("slice", []int{1, 2, 3}, "slice data")

	assert.Equal(nil, getConflictConfigurations(flagSet, fileFlags))

	flagSet.Set("flag1", "2")
	assert.Equal(fmt.Errorf("found conflict flags in command line and config file: from flag: 2 and from config file: 1"),
		getConflictConfigurations(flagSet, fileFlags))

	flagSet.Set("flag2", "1")
	assert.Equal(fmt.Errorf("found conflict flags in command line and config file: from flag: 2 and from config file: 1, from flag: 1 and from config file: 2"),
		getConflictConfigurations(flagSet, fileFlags))
}

func TestGetUnknownFlags(t *testing.T) {
	// TODO
}
