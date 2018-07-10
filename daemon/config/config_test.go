package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) (t *testing.T) {
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
	assert := assert.New(t)
	config := Config{Labels: []string{"a=b", "c=d"}}
	origin := config.Validate()
	assert.Nil(origin)
	config = Config{Labels: []string{"a=b", "cd"}}
	origin = config.Validate()
	assert.Equal(origin, fmt.Errorf("daemon label cd must be in format of key=value"))
	config = Config{Labels: []string{"a="}}
	origin = config.Validate()
	assert.Equal(origin, fmt.Errorf("key and value in daemon label a= cannot be empty"))
}

func TestGetConflictConfigurations(t *testing.T) {
	// TODO
}

func TestGetUnknownFlags(t *testing.T) {
	// TODO
}
