package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spf13/pflag"
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

	fileflags := map[string]interface{}{
		"a": "a1",
		"b": []string{"b1", "b2"},
	}

	flags := pflag.NewFlagSet("cmflags", pflag.ContinueOnError)

	// Test No Flags
	assert.Equal(nil, getConflictConfigurations(flags, fileflags))

	flags.String("c", "c1", "c")
	// Test No Conflicts
	flags.Parse([]string{"--c=c1"})
	assert.Equal(nil, getConflictConfigurations(flags, fileflags))

	// Test Ignore Conflict of Type "Slice"
	flags.StringSlice("b", []string{"b1", "b2"}, "b")
	flags.Parse([]string{"--b=b1,b2"})
	assert.Equal(nil, getConflictConfigurations(flags, fileflags))

	// Test Conflict
	flags.String("a", "a1", "a")
	flags.Parse([]string{"--a=a1"})
	assert.Equal("found conflict flags in command line and config file: from flag: a1 and from config file: a1",
		getConflictConfigurations(flags, fileflags).Error())
}

func TestGetUnknownFlags(t *testing.T) {
	// TODO
}
