package filters

import (
	"encoding/json"
	"sort"
	"strings"
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

type sortString []string

func (s sortString) Len() int {
	return len(s)
}

func (s sortString) Swap(i, j int) {
	tmp := s[i]
	s[i] = s[j]
	s[j] = tmp
}

func (s sortString) Less(i, j int) bool {
	return strings.Compare(s[i], s[j]) < 0
}

func TestFromURLParam(t *testing.T) {
	f := func(i interface{}) string {
		data, _ := json.Marshal(i)
		return string(data)
	}

	tests := []struct {
		name     string
		params   string
		wantErr  bool
		expected map[string][]string
	}{
		{
			name: "normal successful case",
			params: f(map[string][]string{
				"label": {"a=a", "b=b"},
				"id":    {"id1"},
			}),
			wantErr: false,
			expected: map[string][]string{
				"label": {"a=a", "b=b"},
				"id":    {"id1"},
			},
		},
		{
			name: "normal successful case2",
			params: f(map[string]map[string]bool{
				"label": {"a=a": true, "b=b": true},
				"id":    {"id1": true},
			}),
			wantErr: false,
			expected: map[string][]string{
				"label": {"a=a", "b=b"},
				"id":    {"id1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := FromURLParam(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromURLParam() error = %v, wantErr %v", err, tt.wantErr)
			}

			for k, v := range tt.expected {
				compV, exist := res[k]
				if !exist || len(compV) != len(v) {
					t.Errorf("FromURLParam() return %v, want %v", res, tt.expected)
				}

				sort.Sort(sortString(compV))
				sort.Sort(sortString(v))

				for i := range compV {
					if v[i] != compV[i] {
						t.Errorf("FromURLParam() return %v, want %v", res, tt.expected)
					}
				}
			}

		})
	}
}
