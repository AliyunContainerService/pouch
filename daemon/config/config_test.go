package config

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestIterateTLSConfig(t *testing.T) {
	assert := assert.New(t)

	// test nil map will not cause panic
	config := make(map[string]interface{})
	iterateTLSConfig(nil, config)
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
