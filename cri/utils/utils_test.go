package utils

import (
	"testing"
)

func Test_MatchLabelSelector(t *testing.T) {
	type args struct {
		selector map[string]string
		labels   map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Normal Test",
			args: args{
				selector: map[string]string{
					"a1": "b1",
					"a2": "b2",
				},
				labels: map[string]string{
					"a1": "b1",
					"a2": "b2",
				},
			},
			want: true,
		},
		{
			name: "Uncovered Test",
			args: args{
				selector: map[string]string{
					"a1": "b1",
					"a2": "b2",
				},
				labels: map[string]string{
					"a2": "b2",
				},
			},
			want: false,
		},
		{
			name: "Unmatched Test",
			args: args{
				selector: map[string]string{
					"a1": "b0",
					"a2": "b2",
				},
				labels: map[string]string{
					"a1": "b1",
					"a2": "b2",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchLabelSelector(tt.args.selector, tt.args.labels); got != tt.want {
				t.Errorf("matchLabelSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}
