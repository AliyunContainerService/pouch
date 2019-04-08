package config

import (
	"reflect"
	"testing"
)

func TestNewVolumes(t *testing.T) {
	tests := []struct {
		name string
		args *Volumes
		want *Volumes
	}{
		{
			name: "values != nil",
			args: &Volumes{
				values: &[]string{
					"abc:def",
				},
			},
			want: &Volumes{
				values: &[]string{
					"abc:def",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewVolumes(tt.args)
			if got == nil {
				t.Errorf("want got is not nil")
			}
			if got.values == nil {
				t.Errorf("want values is not nil")
			}
			if tt.args != nil && tt.args.values != nil && !reflect.DeepEqual(*got.values, *tt.want.values) {
				t.Errorf("NewVolumes() = %v, want %v", *got.values, *tt.want.values)
			}
		})
	}
}

func TestVolumesSet(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want *Volumes
	}{
		{
			name: "args = \"\"",
			args: []string{""},
			want: &Volumes{
				values: &[]string{""},
			},
		},
		{

			name: "values = abc",
			args: []string{"abc", "def", "abc"},
			want: &Volumes{
				values: &[]string{"abc", "def"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewVolumes(nil)
			if v == nil {
				t.Errorf("want volumes is nil")
			}
			for _, arg := range tt.args {
				err := v.Set(arg)
				if err != nil {
					t.Errorf("failed to set value, err(%v)", err)
				}
			}

			if !reflect.DeepEqual(*v, *tt.want) {
				t.Errorf("Set() = %v, want %v", *v, *tt.want)
			}
		})
	}
}

func TestVolumesString(t *testing.T) {
	tests := []struct {
		name string
		args *Volumes
		want string
	}{
		{
			name: "args = \"\"",
			args: &Volumes{
				values: &[]string{""},
			},
			want: "[]",
		},
		{

			name: "values = abc",
			args: &Volumes{
				values: &[]string{"abc"},
			},
			want: "[abc]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.String()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVolumesValue(t *testing.T) {
	tests := []struct {
		name string
		args *Volumes
		want *[]string
	}{
		{
			name: "args = \"\"",
			args: &Volumes{
				values: &[]string{""},
			},
			want: &[]string{""},
		},
		{

			name: "values = abc",
			args: &Volumes{
				values: &[]string{"abc"},
			},
			want: &[]string{"abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Value()
			if !reflect.DeepEqual(got, *tt.want) {
				t.Errorf("Value() = %v, want %v", got, *tt.want)
			}
		})
	}
}
