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
	assert := assert.New(t)

	// no error
	unknowns := map[string]interface{}{
		"name_a": "value_a",
	}
	flags := pflag.NewFlagSet("testflags", pflag.ExitOnError)
	flags.String("name_a", "value_a", "help_a")
	assert.Equal(getUnknownFlags(flags, unknowns), nil)

	// name_b not in flagSet
	unknowns = map[string]interface{}{
		"name_a": "value_a",
	}
	flags = pflag.NewFlagSet("testflags", pflag.ExitOnError)
	flags.String("name_b", "value_b", "help_b")
	assert.Equal(getUnknownFlags(flags, unknowns).Error(), "unknown flags: name_a")

	// name_a in flagset, name_b not in flagset
	unknowns = map[string]interface{}{
		"name_a": "value_a",
		"name_b": "value_b",
	}

	flags = pflag.NewFlagSet("testflags", pflag.ExitOnError)
	flags.String("name_a", "value_a", "help_a")
	assert.Equal(getUnknownFlags(flags, unknowns).Error(), "unknown flags: name_b")

	// unknowns is none
	unknowns = map[string]interface{}{}

	flags = pflag.NewFlagSet("testflags", pflag.ExitOnError)
	flags.String("name_a", "value_a", "help_a")
	assert.Equal(getUnknownFlags(flags, unknowns), nil)
}
