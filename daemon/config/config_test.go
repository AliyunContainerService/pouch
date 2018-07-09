package config

import (
	"fmt"
	"github.com/alibaba/pouch/apis/types"
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestValidate(t *testing.T) {
	type args struct {
		cfg *Config
	}
	//all complete
	var cfg1 Config
	cfg1.Labels = []string{"key=value"}
	// len != 2
	var cfg2 Config
	cfg2.Labels = []string{"key"}
	// key empty
	var cfg3 Config
	cfg3.Labels = []string{"=value"}
	// value empty
	var cfg4 Config
	cfg4.Labels = []string{"key="}
	// test runtime
	var cfg5 Config
	cfg5.DefaultRuntime = "test"
	cfg5.Runtimes = map[string]types.Runtime{"test": {"path", []string{"runargs"}}}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{name: "test1", args: args{cfg: &cfg1}, wantErr: nil},
		{name: "test2", args: args{cfg: &cfg2}, wantErr: fmt.Errorf("daemon label %s must be in format of key=value", "key")},
		{name: "test3", args: args{cfg: &cfg3}, wantErr: fmt.Errorf("key and value in daemon label %s cannot be empty", "=value")},
		{name: "test4", args: args{cfg: &cfg4}, wantErr: fmt.Errorf("key and value in daemon label %s cannot be empty", "key=")},
		{name: "test5", args: args{cfg: &cfg5}, wantErr: fmt.Errorf("default runtime %s cannot be re-register", "test")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.cfg.Validate()
			assert.Equal(t, tt.wantErr, err)

		})
	}

}
