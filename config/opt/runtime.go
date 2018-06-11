package opt

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"
)

// Runtime defines runtimes information
type Runtime struct {
	values *map[string]types.Runtime
}

// NewRuntime initials a Runtime struct
func NewRuntime(rts *map[string]types.Runtime) *Runtime {
	if rts == nil {
		rts = &map[string]types.Runtime{}
	}

	if *rts == nil {
		*rts = map[string]types.Runtime{}
	}

	rt := &Runtime{values: rts}
	return rt
}

// Set implement Runtime as pflag.Value interface
func (r *Runtime) Set(val string) error {
	splits := strings.Split(val, "=")
	if len(splits) != 2 || splits[0] == "" || splits[1] == "" {
		return fmt.Errorf("invalid runtime %s, correct format must be runtime=path", val)
	}

	name := splits[0]
	path := splits[1]
	if _, exist := (*r.values)[name]; exist {
		return fmt.Errorf("runtime %s already registers to daemon", name)
	}

	(*r.values)[name] = types.Runtime{Path: path}
	return nil
}

// String implement Runtime as pflag.Value interface
func (r *Runtime) String() string {
	var str []string
	for k := range *r.values {
		str = append(str, fmt.Sprintf("%s", k))
	}

	return fmt.Sprintf("%v", str)
}

// Type implement Runtime as pflag.Value interface
func (r *Runtime) Type() string {
	return "value"
}
