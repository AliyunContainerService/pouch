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

func TestGetConflictConfigurations(t *testing.T) {
	names := []string{"a", "b", "c"}

	//init flagSet with names as Value.Name
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSet.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(name)
	})
	for _, name := range names {
		flagSet.String(name, "test", "test")
		flagSet.Set(name, "true")
	}

	//init flagSet with slice
	flagSetSlice := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSetSlice.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(name)
	})
	for _, name := range names {
		flagSetSlice.StringSlice(name, []string{}, "test")
		flagSetSlice.Set(name, "true")
	}

	//init empty flagSet
	flagSetEmpty := pflag.NewFlagSet("test", pflag.ContinueOnError)

	fileFlags1 := map[string]interface{}{
		"a": "a",
		"b": "b",
		"c": "c",
	}
	fileFlags2 := map[string]interface{}{
		"d": "d",
		"e": "f",
		"h": "h",
	}

	type args struct {
		fileFlags map[string]interface{}
		fs        *pflag.FlagSet
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test1", args: args{fileFlags: fileFlags1, fs: flagSet}, wantErr: true},
		{name: "test2", args: args{fileFlags: fileFlags2, fs: flagSet}, wantErr: false},
		{name: "test3", args: args{fileFlags: fileFlags1, fs: flagSetSlice}, wantErr: false},
		{name: "test4", args: args{fileFlags: nil, fs: flagSetSlice}, wantErr: false},
		{name: "test5", args: args{fileFlags: fileFlags1, fs: flagSetEmpty}, wantErr: false},
	}

	//run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := getConflictConfigurations(tt.args.fs, tt.args.fileFlags)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConflictConfigurations() error = %v", err)
				return
			}
		})
	}
}

func TestGetUnknownFlags(t *testing.T) {
	fileFlags1 := map[string]interface{}{
		"a": "a",
		"b": "b",
		"c": "c",
	}
	fileFlags2 := map[string]interface{}{
		"a": "a",
		"b": "b",
		"d": "d",
	}
	fileFlags3 := map[string]interface{}{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
	}
	fileFlagsEmpty := map[string]interface{}{}

	names := []string{"a", "b", "d"}

	//init flagSet with names as Value.Name
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSet.SortFlags = false
	flagSet.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(name)
	})
	for _, name := range names {
		flagSet.String(name, "test", "test")
		flagSet.Set(name, "true")
	}

	//init empty flagSet
	flagSetEmpty := pflag.NewFlagSet("test", pflag.ContinueOnError)

	type args struct {
		fileFlags map[string]interface{}
		fs        *pflag.FlagSet
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test1", args: args{fileFlags: fileFlags1, fs: flagSet}, wantErr: true},
		{name: "test2", args: args{fileFlags: fileFlags2, fs: flagSet}, wantErr: false},
		{name: "test3", args: args{fileFlags: fileFlags3, fs: flagSet}, wantErr: true},
		{name: "test4", args: args{fileFlags: fileFlagsEmpty, fs: flagSet}, wantErr: false},
		{name: "test5", args: args{fileFlags: fileFlags1, fs: flagSetEmpty}, wantErr: true},
		{name: "test6", args: args{fileFlags: fileFlags2, fs: flagSetEmpty}, wantErr: true},
		{name: "test7", args: args{fileFlags: fileFlags3, fs: flagSetEmpty}, wantErr: true},
	}

	//run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := getUnknownFlags(tt.args.fs, tt.args.fileFlags)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUnknownFlags() error = %v", err)
				return
			}
		})
	}
}
