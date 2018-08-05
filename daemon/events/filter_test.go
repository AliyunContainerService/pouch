package events

import (
	"testing"
	"time"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
)

func TestFilter_Match(t *testing.T) {
	type fields struct {
		filter filters.Args
	}
	type args struct {
		ev types.EventsMessage
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				filter: filters.NewArgs(filters.Arg("type", "container")),
			},
			args: args{
				ev: types.EventsMessage{
					Action: "create",
					Type:   types.EventTypeContainer,
					Time:   time.Now().UTC().Unix(),
					ID:     "asdf",
				},
			},
			want: true,
		},
		{
			name: "test2",
			fields: fields{
				filter: filters.NewArgs(filters.Arg("type", "image")),
			},
			args: args{
				ev: types.EventsMessage{
					Action: "create",
					Type:   types.EventTypeContainer,
					Time:   time.Now().UTC().Unix(),
					ID:     "asdf",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ef := &Filter{
				filter: tt.fields.filter,
			}
			if got := ef.Match(tt.args.ev); got != tt.want {
				t.Errorf("Filter.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
