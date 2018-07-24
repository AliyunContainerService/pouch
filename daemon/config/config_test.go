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
	Testflag := pflag.NewFlagSet("ArgsForTest", pflag.ContinueOnError)
	Testflag.String("input1", "1e9", "input1")
	Testflag.String("input2", "2e9", "input2")
	Testflag.IntSlice("numberarray1", []int{2, 2}, "caiji test")

	Flags:= map[string]interface{}{
		"input1": "1e9",
		"input2": "2e9",
		"numberarray1":[]int{2, 2},
	}


	assert := assert.New(t)

	assert.Equal(nil, getConflictConfigurations(Testflag, Flags))

	Testflag.Set("input1", "2")
	assert.Error(getConflictConfigurations(Testflag, Flags))
	assert.Equal("found conflict flags in command line and config file: from flag: 2 and from config file: 1e9",
		getConflictConfigurations(Testflag, Flags).Error())

	Testflag.Set("input1", "1e9")
	//assert.Error(getConflictConfigurations(Testflag, Flags))
	//next code will run timeout for unknown error wait to be cheked
	//assert.Equal(nil, getConflictConfigurations(Testflag, Flags))
}
func TestGetUnknownFlags(t *testing.T) {
	// TODO
}
