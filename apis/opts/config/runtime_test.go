package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/stretchr/testify/assert"
)

func TestNewRuntime(t *testing.T) {
	assert := assert.New(t)

	for _, r := range []*map[string]types.Runtime{
		nil,
		{},
		{
			"a": {},
			"b": {Path: "foo"},
		},
	} {
		runtime := NewRuntime(r)
		// just test no panic here
		assert.NotEmpty(t, runtime)
	}
}

func TestRuntimeSet(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    error
		wantErr bool
		wantStr string
	}{
		{
			name: "valid runtime format",
			args: args{
				value: "foo=bar",
			},
			wantErr: false,
		},
		{
			name: "invalid runtime format",
			args: args{
				value: "foo=",
			},
			want:    fmt.Errorf("invalid runtime foo=, correct format must be runtime=path"),
			wantErr: true,
		},
		{
			name: "duplicate runtime",
			args: args{
				value: "foo=bar",
			},
			want:    fmt.Errorf("runtime foo already registers to daemon"),
			wantErr: true,
		},
	}

	for _, r := range []*map[string]types.Runtime{
		{
			"a": {},
			"b": {Path: "foo"},
		},
	} {
		runtime := NewRuntime(r)
		assert.NotEmpty(t, runtime)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				rt := runtime.Set(tt.args.value)
				if (rt != nil) != tt.wantErr {
					t.Errorf("Runtime.Set() error = %v, wantErr %v", rt, tt.wantErr)
					return
				}
				if tt.wantErr && !reflect.DeepEqual(rt, tt.want) {
					t.Errorf("Runtime.Set() = %v, want %v", rt, tt.want)
				}
			})
		}
	}
}

func TestRuntimeType(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		args    args
		err     error
		wantErr bool
		wantStr string
	}{
		{
			name:    "get type of Runtime",
			wantErr: false,
			wantStr: "runtime",
		},
	}
	{
		ul := &Runtime{}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				out := ul.Type()
				if !tt.wantErr && !reflect.DeepEqual(out, tt.wantStr) {
					t.Errorf("Runtime.Type() = %v, want %v", out, tt.wantStr)
				}
			})
		}
	}
}
