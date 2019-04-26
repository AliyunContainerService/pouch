package mgr

import (
	"context"
	"regexp"
	"strings"

	"github.com/alibaba/pouch/pkg/log"
	"github.com/alibaba/pouch/pkg/utils/filters"
)

var (
	labelFilter  = "label"
	idFilter     = "id"
	nameFilter   = "name"
	statusFilter = "status"
)

// filterContext includes conditions provide for filter
type filterContext struct {
	condition  map[string][]string
	all        bool
	filterFunc ContainerFilter
}

// newFilterContext initials a filterContext struct, and validate option.Filter
func newFilterContext(option *ContainerListOption) (*filterContext, error) {
	if option == nil {
		return &filterContext{}, nil
	}

	// validate filter here
	if err := filters.Validate(option.Filter); err != nil {
		return nil, err
	}
	return &filterContext{
		condition:  option.Filter,
		all:        option.All,
		filterFunc: option.FilterFunc,
	}, nil
}

// matchFilter filters value matchs field of condition.
func (fc *filterContext) matchFilter(field, value string) bool {
	filters, exist := fc.condition[field]
	if !exist {
		// return true if field is not exist
		return true
	}

	if value == "" || len(filters) == 0 {
		return false
	}

	var (
		match = false
		err   error
	)
	for _, f := range filters {
		if match, err = regexp.MatchString(f, value); err != nil {
			continue
		}
		if match {
			break
		}
	}

	return match
}

// matchKVFilter filters map value matchs field of condition, also support
// filter not equal, as for unequal condition `key=value`, we filter `$k=$v`
// which $k equal key and $v not equal value.
func (fc *filterContext) matchKVFilter(field string, value map[string]string) bool {
	filters, exist := fc.condition[field]
	if !exist {
		// return true if field is not exist
		return true
	}

	if len(filters) == 0 {
		return false
	}

	// repeat value not support, eg: {"label": []string{"a=b", "a=c"}}, c will overwrite b
	equalKV := make(map[string]string)
	unequalKV := make(map[string]string)
	for _, f := range filters {
		if f == "" {
			continue
		}
		splits := strings.SplitN(f, "!=", 2)
		if len(splits) == 2 {
			unequalKV[splits[0]] = splits[1]
			continue
		}

		splits = strings.SplitN(f, "=", 2)
		if len(splits) == 2 {
			equalKV[splits[0]] = splits[1]
			continue
		}

		equalKV[f] = ""
	}

	if len(equalKV) == 0 && len(unequalKV) == 0 {
		return false
	}

	// filter equal condition
	for k, equalValue := range equalKV {
		// if not find equal (k, v) pair, return false
		if v, exist := value[k]; !exist || (equalValue != "" && equalValue != v) {
			return false
		}
	}

	// filter unequal condition
	for k, unequalValue := range unequalKV {
		// if key not exist or pair (k, v) found, return false
		if v, exist := value[k]; exist && unequalValue != v {
			continue
		}

		return false
	}

	return true
}

// filter does all select container work.
func (fc *filterContext) filter(c *Container) bool {
	isNonStop := c.IsRunningOrPaused()

	// if filterFunc is defined, we not try other filter condition
	if fc.filterFunc != nil {
		if fc.filterFunc(c) {
			// consider if need to add non-running container
			return isNonStop || fc.all
		}
		return false
	}

	if len(fc.condition) == 0 {
		// if no filter condition, skip filter process,
		// add container according to flag `all` or running status
		return isNonStop || fc.all
	}

	// filter must meets all condition. skip loop when condition not match
	var match, statusKey = true, false
	for name := range fc.condition {
		if _, exist := fc.condition[name]; !exist {
			// jump filter if field is not exist
			continue
		}

		switch name {
		case labelFilter:
			match = fc.matchKVFilter(labelFilter, c.Config.Labels)
		case idFilter:
			match = fc.matchFilter(idFilter, c.ID)
		case nameFilter:
			match = fc.matchFilter(nameFilter, c.Name)
		case statusFilter:
			match = fc.matchFilter(statusFilter, string(c.State.Status))
		default:
			continue
		}

		if !match {
			// do not return in loop
			break
		}

		if name == statusFilter {
			statusKey = true
		}
	}

	if match && !statusKey {
		// consider if need to add non-running container
		match = isNonStop || fc.all
	}

	return match
}

// List returns the container's list.
func (mgr *ContainerManager) List(ctx context.Context, option *ContainerListOption) ([]*Container, error) {
	var cons []*Container
	list := mgr.cache.Values(nil)

	fc, err := newFilterContext(option)
	if err != nil {
		return nil, err
	}

	for id, obj := range list {
		c, ok := obj.(*Container)
		if !ok {
			log.With(ctx).Warningf("getting container list, drop partial container cache %s", id)
			continue
		}

		if fc.filter(c) {
			cons = append(cons, c)
		}
	}

	return cons, nil
}
