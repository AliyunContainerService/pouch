package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/stretchr/testify/assert"
)

func TestNewRuntime(t *testing.T) {
	type args struct {
		val *map[string]types.Runtime
	}

	//create a nil map that makes *rts = nil
	var nilmap = map[string]types.Runtime(nil)

	tests := []struct {
		name string
		args args
		want *Runtime
	}{
		{
			name: "rts = nil",
			args: args{
				val: nil,
			},
			want: &Runtime{
				values: &map[string]types.Runtime{},
			},
		},
		{
			name: "*rts = nil",
			args: args{
				val: &nilmap,
			},
			want: &Runtime{
				values: &map[string]types.Runtime{},
			},
		},
		{
			name: "rts is valid",
			args: args{
				val: &map[string]types.Runtime{
					"runtime_name1": {},
					"runtime_name2": {Path: "$PATH"},
				},
			},
			want: &Runtime{
				values: &map[string]types.Runtime{
					"runtime_name1": {},
					"runtime_name2": {Path: "$PATH"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRuntime(tt.args.val)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRuntime() = %v, want %v", got, tt.want)
			}
		})
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

func TestRuntimeString(t *testing.T) {
	type fields struct {
		values *map[string]types.Runtime
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "get string of Runtime with one element",
			fields: fields{
				values: &(map[string]types.Runtime{
					"runtime_name": {Path: "$PATH"},
				}),
			},
			want: "[runtime_name]",
		},
		{
			name: "get string of Runtime with nil",
			fields: fields{
				values: &(map[string]types.Runtime{}),
			},
			want: "[]",
		},
		{
			name: "get string of Runtime with empty string",
			fields: fields{
				values: &(map[string]types.Runtime{
					"": {Path: "$PATH"},
				}),
			},
			want: "[]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Runtime{
				values: tt.fields.values,
			}
			if got := r.String(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Runtime.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
