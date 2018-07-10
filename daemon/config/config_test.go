package config

import (
	"testing"
	"fmt"

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

	// test1 daemon label %s must be in format of key=value
	listen1 := []string{"1","2","3"}
	labels1 := []string{"k1"}
	
	// mock data
	cfg1 := &Config{
		Listen : listen1,
		Labels : labels1,
	}

	err1 := cfg1.Validate()

	assert.Equal(t, fmt.Errorf("daemon label k1 must be in format of key=value"), err1)

	// test2 key and value in daemon label %s cannot be empty
	listen2 := []string{"1","2","3"}
	labels2 := []string{"="}
	
	// mock data
	cfg2 := &Config{
		Listen : listen2,
		Labels : labels2,
	}

	err2 := cfg2.Validate()

	assert.Equal(t, fmt.Errorf("key and value in daemon label = cannot be empty"), err2)

	// test3 success
	listen3 := []string{"1","2","3"}
	labels3 := []string{"k1=v1","k2=v2","k3=v3"}
	
	// mock data
	cfg3 := &Config{
		Listen : listen3,
		Labels : labels3,
	}

	err3 := cfg3.Validate()

	assert.Equal(t, nil, err3)
}

func TestGetConflictConfigurations(t *testing.T) {
	// TODO
}

func TestGetUnknownFlags(t *testing.T) {

	// Test2 unknown flags: %s
	flagSet2 := &pflag.FlagSet{
	}

	map2 := map[string]interface{}{
	}

	err2 := getUnknownFlags(flagSet2, map2)

	assert.Equal(t, nil, err2)
}
