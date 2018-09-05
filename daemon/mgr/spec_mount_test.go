package mgr

import (
	"reflect"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func Test_sortMounts(t *testing.T) {
	tests := []struct {
		name string
		args []specs.Mount
		want []specs.Mount
	}{
		{
			"Mounts is nil case",
			nil,
			nil,
		},
		{
			"Normal Mounts",
			[]specs.Mount{
				{Destination: "/etc/resolv.conf"},
				{Destination: "/etc"},
			},
			[]specs.Mount{
				{Destination: "/etc"},
				{Destination: "/etc/resolv.conf"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortMounts(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortMounts() = %v, want %v", got, tt.want)
			}
		})
	}
}
