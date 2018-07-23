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

	// test1. name in flagset
	unknowns := map[string]interface{}{
		"name": "Anthony",
	}
	f := pflag.NewFlagSet("test0", pflag.ContinueOnError)
	f.String("name", "Anthony", "String Test")

	res := getUnknownFlags(f, unknowns)

	check := 0
	if res != nil {
		check = 1
	}
	assert.Equal(check, 0)

	// test2. sex not in flagset

	unknowns = map[string]interface{}{
		"name": "Anthony",
	}
	f = pflag.NewFlagSet("test1", pflag.ContinueOnError)
	f.String("sex", "Female", "Female")

	res = getUnknownFlags(f, unknowns)

	check = 0
	if res != nil {
		check = 1
	}
	assert.Equal(check, 1)

	// test3. one in flagset,another not in
	unknowns = map[string]interface{}{
		"name": "jachin",
		"sex":  "Male",
	}

	f = pflag.NewFlagSet("test2", pflag.ContinueOnError)
	f.String("name", "jachin", "jachin")

	res = getUnknownFlags(f, unknowns)
	check = 0
	if res != nil {
		check = 1
	}
	assert.Equal(check, 1)
	// test4. noone in flagset
	unknowns = map[string]interface{}{}

	f = pflag.NewFlagSet("test3", pflag.ContinueOnError)
	f.String("name", "jachin", "jachin")

	res = getUnknownFlags(f, unknowns)

	check = 0
	if res != nil {
		check = 1
	}
	assert.Equal(check, 0)
}
