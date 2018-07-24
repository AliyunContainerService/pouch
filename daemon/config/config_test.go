package config

import (
	"fmt"
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

	assert := assert.New(t)

	fileFlags := map[string]interface{}{
		"fileFlagsName1": "fileFlagsValue1",
		"fileFlagsName2": "fileFlagsValue2",
		"slice":          "sliceData",
		"commonName":     "fileFlagsCommonValue",
	}

	flagSet := pflag.NewFlagSet("FlagConfig", pflag.ContinueOnError)
	flagSet.String("flagSetName1", "", "flagSetName1")
	flagSet.String("flagSetName2", "", "flagSetName2")
	flagSet.String("commonName", "", "commonName")
	flagSet.IntSlice("slice", []int{1, 2, 3}, "sliceData")

	/* test for slice type. even if there is common key, it will return nil */
	flagSet.Parse([]string{"--slice=1,2,3"})
	assert.Equal(nil, getConflictConfigurations(flagSet, fileFlags))

	/* test for different key, it will return nil */
	flagSet.Set("flagSetName1", "flagSetValue1")
	flagSet.Set("flagSetName2", "flagSetValue2")
	assert.Equal(nil, getConflictConfigurations(flagSet, fileFlags))

	/* test for common key, it will return string that has conflict key */
	flagSet.Set("commonName", "flagSetCommonValue")
	assert.Equal(fmt.Errorf("found conflict flags in command line and config file: from flag: flagSetCommonValue and from config file: fileFlagsCommonValue"),
		getConflictConfigurations(flagSet, fileFlags))
}

func TestGetUnknownFlags(t *testing.T) {
	// TODO

}
