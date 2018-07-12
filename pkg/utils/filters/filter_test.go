package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFilter(t *testing.T) {
	assert := assert.New(t)
	type tCase struct {
		filter   []string
		ok       bool
		errorMsg string
	}

	for _, t := range []tCase{
		{
			filter: []string{"id=a", "name=b"},
			ok:     true,
		},
		{
			filter: []string{"status=running"},
			ok:     true,
		},
		{
			filter:   []string{"foo"},
			ok:       false,
			errorMsg: "Bad format of filter, expected name=value",
		},
		{
			filter:   []string{"foo=bar"},
			ok:       false,
			errorMsg: "Invalid filter",
		},
		{
			filter: []string{"label=a", "label=a=b"},
			ok:     true,
		},
		{
			filter: []string{"label=a!=b", "id=aaa"},
			ok:     true,
		},
	} {
		_, err := Parse(t.filter)
		if t.ok {
			assert.NoError(err)
		} else {
			assert.Contains(err.Error(), t.errorMsg)
		}
	}
}

func TestValidate(t *testing.T) {
	type args struct {
		filter map[string][]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal successful case",
			args: args{
				filter: map[string][]string{
					"id": {"a"},
				},
			},
			wantErr: false,
		},
		{
			name: "failure case with nonexistence",
			args: args{
				filter: map[string][]string{
					"id":           {"a"},
					"nonexistence": {"b"},
				},
			},
			wantErr: true,
		},
		{
			name: "failure case with future filter name",
			args: args{
				filter: map[string][]string{
					"id":     {"a"},
					"before": {"b"}, // this will be implemented in the future.
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.args.filter); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
