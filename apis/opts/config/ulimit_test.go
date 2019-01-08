package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/apis/types"
	units "github.com/docker/go-units"
	"github.com/stretchr/testify/assert"
)

func TestUlimitSet(t *testing.T) {
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
			name: "valid ulimit format 1",
			args: args{
				val: "nofile=512:1024",
			},
			wantErr: false,
		},
		{
			name: "valid ulimit format 2",
			args: args{
				val: "nofile=1024",
			},
			wantErr: false,
		},
		{
			name: "valid ulimit format 3",
			args: args{
				val: "cpu=2:4",
			},
			wantErr: false,
		},
		{
			name: "valid ulimit format 4",
			args: args{
				val: "cpu=6",
			},
			wantErr: false,
		},
		{
			name: "invalid ulimit value type 1",
			args: args{
				val: "",
			},
			wantErr: true,
			err:     fmt.Errorf("invalid ulimit argument: "),
		},
		{
			name: "invalid ulimit value type 2",
			args: args{
				val: "nofile=512:1024:2048",
			},
			wantErr: true,
			err:     fmt.Errorf("too many limit value arguments - 512:1024:2048, can only have up to two, `soft[:hard]`"),
		},
		{
			name: "the hard ulimit value is less than soft",
			args: args{
				val: "nofile=1024:1",
			},
			wantErr: true,
			err:     fmt.Errorf("ulimit soft limit must be less than or equal to hard limit: 1024 > 1"),
		},
		{
			name: "bad ulimit format",
			args: args{
				val: "cpu:512:1024",
			},
			wantErr: true,
			err:     fmt.Errorf("invalid ulimit argument: cpu:512:1024"),
		},
		{
			name: "invalid ulimit type",
			args: args{
				val: "foo=1024:1024",
			},
			wantErr: true,
			err:     fmt.Errorf("invalid ulimit type: foo"),
		},
	}
	{
		ul := &Ulimit{}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ul.Set(tt.args.val)
				if (err != nil) != tt.wantErr {
					t.Errorf("Ulimit.Set() error = %v, wantErr %v", err, tt.wantErr)
				}
				if (tt.err != nil) && !reflect.DeepEqual(err, tt.err) {
					t.Errorf("Ulimit.Set() = %v, want %v", err, tt.err)
				}
			})
		}
	}
}

func TestUlimitType(t *testing.T) {
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
			name:    "get type of Ulimit",
			wantErr: false,
			wantStr: "ulimit",
		},
	}
	{
		ul := &Ulimit{}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				out := ul.Type()
				if !tt.wantErr && !reflect.DeepEqual(out, tt.wantStr) {
					t.Errorf("Ulimit.Type() = %v, want %v", out, tt.wantStr)
				}
			})
		}
	}
}

func TestUlimitString(t *testing.T) {
	type args struct {
		values map[string]*units.Ulimit
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"get string of Ulimit with valid element",
			args{
				values: map[string]*units.Ulimit{
					"a": {
						Name: "core",
						Hard: 5,
						Soft: 3,
					},
				},
			},
			"[core=3:5]",
		},
		{
			"get string of Runtime with nil",
			args{
				values: map[string]*units.Ulimit{},
			},
			"[]",
		},
		{
			"get string of Runtime with empty Ulimit",
			args{
				values: map[string]*units.Ulimit{
					"a": {},
				},
			},
			"[=0:0]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ul := &Ulimit{
				values: tt.args.values,
			}
			if got := ul.String(); got != tt.want {
				t.Errorf("Ulimit.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUlimitValue(t *testing.T) {
	assert := assert.New(t)
	type args struct {
		values map[string]*units.Ulimit
	}
	tests := []struct {
		name string
		args args
		want types.Ulimit
	}{
		{
			name: "get all values as type Ulimit",
			want: types.Ulimit{
				Name: "nofile",
				Hard: int64(1024),
				Soft: int64(512),
			},
			args: args{
				values: map[string]*units.Ulimit{
					"a": {Name: "nofile", Hard: int64(1024), Soft: int64(512)},
				},
			},
		},
	}
	{
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ul := Ulimit{values: tt.args.values}
				got := ul.Value()
				assert.NotEmpty(t, got)
				for _, v := range got {
					if !reflect.DeepEqual(*v, tt.want) {
						t.Errorf("ul.value() = %v, want %v", *v, tt.want)
					}
				}
			})
		}
	}
}
