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

	fileflags := map[string]interface{}{
		"name_a": "value_a",
		"name_b": "value_b",
		"name_s": []string{"value_s1", "value_s2"},
	}

	flags := pflag.NewFlagSet("testflags", pflag.ExitOnError)
	flags.String("name_c", "value_c", "help_c")

	// While flags is not set
	assert.Equal(nil, getConflictConfigurations(flags, fileflags))

	// While no conflict
	flags.Parse([]string{"--name_c=value_c"})
	assert.Equal(nil, getConflictConfigurations(flags, fileflags))

	// Create conflict
	flags.String("name_a", "value_a", "help_a")
	flags.String("name_b", "value_b", "help_b")
	flags.Parse([]string{"--name_a=value_a", "--name_b=value_b"})
	assert.Equal("found conflict flags in command line and config file: from flag: value_a and from config file: value_a, from flag: value_b and from config file: value_b",
		getConflictConfigurations(flags, fileflags).Error())

	// Add slice conflict, it should be ignored
	flags.StringSlice("name_s", []string{"value_s1", "value_s2"}, "help_s")
	flags.Parse([]string{"--name_s=value_s1,value_s2"})
	assert.Equal("found conflict flags in command line and config file: from flag: value_a and from config file: value_a, from flag: value_b and from config file: value_b",
		getConflictConfigurations(flags, fileflags).Error())
}

func TestGetUnknownFlags(t *testing.T) {

}
