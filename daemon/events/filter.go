package events

import (
	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
)

// Filter uses to filter out pouch events from a stream
type Filter struct {
	filter filters.Args
}

// NewFilter initializes a new Filter.
func NewFilter(filter filters.Args) *Filter {
	return &Filter{filter: filter}
}

// Match returns true when the event ev is included by the filters
func (ef *Filter) Match(ev types.EventsMessage) bool {
	// TODO(ziren): add more filters
	return ef.filter.ExactMatch("event", ev.Action) &&
		ef.filter.ExactMatch("type", string(ev.Type))
}
