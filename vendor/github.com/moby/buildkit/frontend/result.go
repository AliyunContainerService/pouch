package frontend

import "github.com/moby/buildkit/solver"

type Result struct {
	Ref      solver.CachedResult
	Refs     map[string]solver.CachedResult
	Metadata map[string][]byte
}

func (r *Result) EachRef(fn func(solver.CachedResult) error) (err error) {
	if r.Ref != nil {
		err = fn(r.Ref)
	}
	for _, r := range r.Refs {
		if r != nil {
			if err1 := fn(r); err1 != nil && err == nil {
				err = err1
			}
		}
	}
	return err
}
