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
	flags := pflag.NewFlagSet("case0", pflag.ContinueOnError)
	flags.String("a", "b", "c")
	type args struct {
		flagSet   *pflag.FlagSet
		fileFlags map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{name: "test0", args: args{flagSet: flags, fileFlags: map[string]interface{}{"a": "a"}}, want: nil},
		{name: "test1", args: args{flagSet: flags, fileFlags: map[string]interface{}{"a": "b"}}, want: nil},
		{name: "test2", args: args{flagSet: flags, fileFlags: map[string]interface{}{"a": "c"}}, want: nil},
		{name: "test3", args: args{flagSet: flags, fileFlags: map[string]interface{}{"a": "d"}}, want: nil},
		{name: "test4", args: args{flagSet: flags, fileFlags: map[string]interface{}{"b": "a"}}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getUnknownFlags(tt.args.flagSet, tt.args.fileFlags)
			if got != tt.want {
				t.Errorf("!getUnknownFlags = %v, want %v", got, tt.want)
				return
			}
		})
	}
}
