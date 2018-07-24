package config

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"reflect"
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
	Testflag.String("input1", "1e9", "flag1")
	Testflag.String("input2", "2e9", "flag2")
	Testflag.IntSlice("numberarray1", []int{2, 2}, "caiji test")

	Flags := map[string]interface{}{
		"input1":      "1e9",
		"input2":      "2e9",
		"numberarray": []int{1, 2},
	}

	assert := assert.New(t)

	assert.Equal(nil, getConflictConfigurations(Testflag, Flags))

	Testflag.Set("input1", "2")
	assert.Error(getConflictConfigurations(Testflag, Flags))
	assert.Equal("found conflict flags in command line and config file: from flag: 2 and from config file: 1e9",
		getConflictConfigurations(Testflag, Flags).Error())

}

func TestGetUnknownFlags(t *testing.T) {
	assert := assert.New(t)
	type args struct {
		flagSet *pflag.FlagSet
		fileFlags map[string]interface{}
	}
	fileFlags := map[string]interface{}{
		"k1": "v1",
		"k2": "v2",
	}
	flagSet := pflag.NewFlagSet("TestConfig", pflag.ContinueOnError)
	flagSet.String("k1", "v1", "k1")
	flagSet.String("k2", "v2", "k2")

	fileFlags2 := map[string]interface{}{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	}
	tests := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{flagSet:flagSet,fileFlags:fileFlags}, want: nil, wantErr: false},
		{name: "test2", args: args{flagSet:flagSet,fileFlags:fileFlags2}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := getUnknownFlags(tt.args.flagSet,tt.args.fileFlags)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUnknownFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(err, tt.want) {
				assert.Equal("unknown flags: k3", err.Error())
			}
		})
	}
}

