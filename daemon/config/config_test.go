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
	//without  conflict
	flagSet0 := pflag.NewFlagSet("flagset1", 0)
	flagSet0.String("a", "1", "test")
	s0 := []string{"--a", "31111"}
	flagSet0.Parse(s0)
	//with slice conflict
	flagSet1 := pflag.NewFlagSet("flagset1", 0)
	flagSet1.StringSlice("slice", []string{"111", "222", "333"}, "test")
	s1 := []string{"--slice=aaa,bbb,ccc"}
	flagSet1.Parse(s1)
	//with conflict
	flagSet2 := pflag.NewFlagSet("flagset2", 0)
	flagSet2.String("b", "3", "test")
	s2 := []string{"--b", "222"}
	flagSet2.Parse(s2)
	fileFlags := map[string]interface{}{
		"b":     "2",
		"slice": []string{"111", "222", "333"},
	}
	error := getConflictConfigurations(flagSet0, fileFlags)
	if error != nil {
		t.Fatal(error)
	}
	error = getConflictConfigurations(flagSet1, fileFlags)
	if error != nil {
		t.Fatal(error)
	}
	error = getConflictConfigurations(flagSet2, fileFlags)
	if error == nil {
		t.Fatal("expect get driver not found error, but err is nil")
	}

}

func TestGetUnknownFlags(t *testing.T) {
	// TODO
}
