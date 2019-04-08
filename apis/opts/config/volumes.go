package config

import (
	"fmt"
)

// Volumes holds a list of values.
type Volumes struct {
	values *[]string
}

// NewVolumes creates a new Volumes.
func NewVolumes(v *Volumes) *Volumes {
	var values []string
	if v == nil {
		v = &Volumes{}
	}
	if v.values == nil {
		v.values = &values
	}
	return v
}

// Set implement Volumes as pflag.Value interface.
func (v *Volumes) Set(val string) error {
	for _, s := range *v.values {
		if s == val {
			return nil
		}
	}
	(*v.values) = append((*v.values), val)
	return nil
}

// String implement Volumes as pflag.Value interface.
func (v *Volumes) String() string {
	return fmt.Sprintf("%v", *v.values)
}

// Type implement Volumes as pflag.Value interface.
func (v *Volumes) Type() string {
	return "volumes"
}

// Value return values as type Volumes
func (v *Volumes) Value() []string {
	return (*v.values)
}
