package config

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"

	units "github.com/docker/go-units"
)

// Ulimit defines ulimit options.
type Ulimit struct {
	values map[string]*units.Ulimit
}

// Set implement Ulimit as pflag.Value interface.
func (u *Ulimit) Set(val string) error {
	ul, err := units.ParseUlimit(val)
	if err != nil {
		return err
	}

	if u.values == nil {
		u.values = make(map[string]*units.Ulimit)
	}

	u.values[ul.Name] = ul
	return nil
}

// String implement Ulimit as pflag.Value interface.
func (u *Ulimit) String() string {
	var str []string
	for _, ul := range u.values {
		str = append(str, ul.String())
	}

	return fmt.Sprintf("%v", str)
}

// Type implement Ulimit as pflag.Value interface.
func (u *Ulimit) Type() string {
	return "ulimit"
}

// Value return ulimit values as type Ulimit
func (u *Ulimit) Value() []*types.Ulimit {
	var ulimit []*types.Ulimit
	for _, ul := range u.values {
		ulimit = append(ulimit, &types.Ulimit{
			Name: ul.Name,
			Hard: ul.Hard,
			Soft: ul.Soft,
		})
	}

	return ulimit
}
