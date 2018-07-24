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

	//without unknown flags
	flagSet := pflag.NewFlagSet("flagset1", 0)
	flagSet.String("a", "1", "test")
	flagSet.String("b", "2", "test")
	//with unknown flags
	flagSet2 := pflag.NewFlagSet("flagset2", 0)
	flagSet2.String("a", "1", "test")
	flagSet2.String("c", "2", "test")
	fileFlags := map[string]interface{}{
		"a": "1",
		"b": "2",
	}
	error := getUnknownFlags(flagSet, fileFlags)
	if error != nil {
		t.Fatal(error)
	}
	error = getUnknownFlags(flagSet2, fileFlags)
	if error == nil {
		t.Fatal("expect get driver not found error, but err is nil")
	}
}
