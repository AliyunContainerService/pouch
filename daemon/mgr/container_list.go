package mgr

import (
	"context"
	"fmt"
)

// the filter tags set allowed when pouch ps -f
var acceptedContainerFilterTags = map[string]bool{
	"label":  true,
	"id":     true,
	"name":   true,
	"status": true,
	/*
		// TODO(huamin.thm): the following list key should also support
		"before":  true,
		"since":   true,
		"exited":  true,
		"volume":  true,
		"network": true,
	*/
}

// List returns the container's list.
func (mgr *ContainerManager) List(ctx context.Context, opt *ContainerListOption) ([]*Container, error) {
	filter := opt.Filter
	if err := filter.Validate(acceptedContainerFilterTags); err != nil {
		return nil, err
	}

	var cons []*Container

	list, err := mgr.Store.List()
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	labelFilter := filter.Get("label")
	idFilter := filter.Get("id")
	nameFilter := filter.Get("name")
	statusFilter := filter.Get("status")

	for _, obj := range list {
		c, ok := obj.(*Container)
		if !ok {
			return nil, fmt.Errorf("failed to get container list, invalid meta type")
		}

		isNonStop := c.IsRunningOrPaused()
		if opt.FilterFunc != nil {
			if opt.FilterFunc(c) {
				if isNonStop || opt.All {
					cons = append(cons, c)
				}
			}
			continue
		}

		match := true
		if len(labelFilter) != 0 {
			match = match && filter.MatchKVList("label", c.Config.Labels)
		}
		if len(idFilter) != 0 {
			match = match && filter.Match("id", c.ID)
		}
		if len(nameFilter) != 0 {
			match = match && filter.Match("name", c.Name)
		}
		if len(statusFilter) != 0 {
			match = match && filter.Match("status", string(c.State.Status))
		}
		if match && len(statusFilter) == 0 {
			match = isNonStop || opt.All
		}
		if match {
			cons = append(cons, c)
		}
	}

	return cons, nil
}
