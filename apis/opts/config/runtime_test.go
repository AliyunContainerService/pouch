package config

import (
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/stretchr/testify/assert"
)

func TestNewRuntime(t *testing.T) {
	assert := assert.New(t)

	var tmp = map[string]types.Runtime(nil)

	for _, r := range []*map[string]types.Runtime{
		nil,
		&tmp,
		{
			"a": {},
			"b": {Path: "foo"},
		},
	} {
		runtime := NewRuntime(r)
		// just test no panic here
		assert.NoError(runtime.Set("foo=bar"))
	}
}

func TestRuntime_Set(t *testing.T) {
	type fields struct {
		values *map[string]types.Runtime
	}
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "incorrect format that the number of splits != 3",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "runtime_name $PATH",
			},
			wantErr: true,
		},
		{
			name: "incorrect format that 1st split = \"\"",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "runtime_name=",
			},
			wantErr: true,
		},
		{
			name: "incorrect format that 2nd split = \"\"",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "=$PATH",
			},
			wantErr: true,
		},
		{
			name: "set one element already registers to daemon",
			fields: fields{
				values: &(map[string]types.Runtime{
					"runtime_name": {Path: "$PATH"},
				}),
			},
			args: args{
				val: "runtime_name=$PATH",
			},
			wantErr: true,
		},
		{
			name: "set one valid element",
			fields: fields{
				values: &(map[string]types.Runtime{
					"runtime_name": {Path: "$PATH"},
				}),
			},
			args: args{
				val: "runtime_name2=$PATH",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Runtime{
				values: tt.fields.values,
			}
			if err := r.Set(tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("Runtime.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRuntime_String(t *testing.T) {
	type fields struct {
		values *map[string]types.Runtime
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				values: &(map[string]types.Runtime{
					"runtime_name": {Path: "$PATH"},
				}),
			},
			want: "[runtime_name]",
		},
		{
			name: "",
			fields: fields{
				values: &(map[string]types.Runtime{
					"runtime_name1": {Path: "$PATH"},
					"runtime_name2": {Path: "$PATH"},
				}),
			},
			want: "[runtime_name1 runtime_name2]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Runtime{
				values: tt.fields.values,
			}
			if got := r.String(); got != tt.want {
				t.Errorf("Runtime.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuntime_Type(t *testing.T) {
	type fields struct {
		values *map[string]types.Runtime
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "implement Runtime as pflag.Value interface",
			fields: fields{
				values: &(map[string]types.Runtime{
					"runtime_name": {Path: "$PATH"},
				}),
			},
			want: "runtime",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Runtime{
				values: tt.fields.values,
			}
			if got := r.Type(); got != tt.want {
				t.Errorf("Runtime.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}
