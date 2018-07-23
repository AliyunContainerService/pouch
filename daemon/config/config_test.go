package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spf13/pflag"

	// "fmt"

	// "strings"

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
	// 'Baiji' Group 1, Zeyu Sun
	assert := assert.New(t)
	// Command Line Flags
	commandLineFlags := pflag.NewFlagSet("commandLineFlags", pflag.ContinueOnError)
	// Adding Flags
	commandLineFlags.String("f1", "1", "first String Flag")
	commandLineFlags.String("f2", "", "second String Flag")
	commandLineFlags.IntSlice("f3", []int{}, "slice Type Flag")

	// Config File
	configFileFlags := map[string]interface{}{
		// "f1" : "2",
	}

	// commandLineFlags.Visit(func(f *pflag.Flag) {
	// 	flagType := f.Value.Type()
	// 	if strings.Contains(flagType, "Slice") {
	// 		fmt.Println("SLICE")
	// 		return
	// 	}

	// })

	/* 4. If commandLineFlagsSet contains 'Slice' type */
	commandLineFlags.Set("f3", "[]")
	configFileFlags["f3"] = "1"
	assert.Equal(nil, getConflictConfigurations(commandLineFlags, configFileFlags))

	commandLineFlags.Set("f1", "2")
	configFileFlags["f1"] = "2"
	// fmt.Println(commandLineFlags.Lookup("f1").Value.Type())
	/* 1. With Same Flag name(Same Value), Conflict */
	assert.Error(getConflictConfigurations(commandLineFlags, configFileFlags))
	assert.Equal(getConflictConfigurations(commandLineFlags, configFileFlags).Error(),
	"found conflict flags in command line and config file: from flag: 2 and from config file: 2")

	/* 2. With Same Flag name(Different Value), Conflict */
	commandLineFlags.Set("f1", "1")
	assert.Error(getConflictConfigurations(commandLineFlags, configFileFlags))
	assert.Equal(getConflictConfigurations(commandLineFlags, configFileFlags).Error(),
	"found conflict flags in command line and config file: from flag: 1 and from config file: 2")

	/* 3. Two Configs's flag Intersects */
	commandLineFlags.Set("f2", "0")
	configFileFlags["f4"] = "3"
	assert.Error(getConflictConfigurations(commandLineFlags, configFileFlags))
	assert.Equal(getConflictConfigurations(commandLineFlags, configFileFlags).Error(),
	"found conflict flags in command line and config file: from flag: 1 and from config file: 2")
	
}

func TestGetUnknownFlags(t *testing.T) {
	// TODO
}
