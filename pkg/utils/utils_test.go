package utils

import (
	goerrors "errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type tCase struct {
	name     string
	input    int64
	expected string
	err      error
}

func TestFormatSize(t *testing.T) {
	assert := assert.New(t)
	kvs := map[int64]string{
		-1024:         "0.00 B",
		0:             "0.00 B",
		1024:          "1.00 KB",
		1024000:       "1000.00 KB",
		1024000000000: "953.67 GB",
	}

	for k, v := range kvs {
		size := FormatSize(k)
		assert.Equal(v, size)
	}
}

func TestFormatTimeInterval(t *testing.T) {

	for _, tc := range []tCase{
		{
			name:     "second",
			input:    time.Now().Add(0 - Second).UnixNano(),
			expected: "1 second",
			err:      nil,
		}, {
			name:     "minute",
			input:    time.Now().Add(0 - Minute).UnixNano(),
			expected: "1 minute",
			err:      nil,
		}, {
			name:     "hour",
			input:    time.Now().Add(0 - Hour).UnixNano(),
			expected: "1 hour",
			err:      nil,
		}, {
			name:     "day",
			input:    time.Now().Add(0 - Day).UnixNano(),
			expected: "1 day",
			err:      nil,
		}, {
			name:     "week",
			input:    time.Now().Add(0 - Week).UnixNano(),
			expected: "1 week",
			err:      nil,
		}, {
			name:     "month",
			input:    time.Now().Add(0 - Month).UnixNano(),
			expected: "1 month",
			err:      nil,
		}, {
			name:     "year",
			input:    time.Now().Add(0 - Year).UnixNano(),
			expected: "1 year",
			err:      nil,
		},
		{
			name:     "seconds",
			input:    time.Now().Add(0 - Second*3).UnixNano(),
			expected: "3 seconds",
			err:      nil,
		}, {
			name:     "minutes",
			input:    time.Now().Add(0 - Minute*3).UnixNano(),
			expected: "3 minutes",
			err:      nil,
		}, {
			name:     "hours",
			input:    time.Now().Add(0 - Hour*3).UnixNano(),
			expected: "3 hours",
			err:      nil,
		}, {
			name:     "days",
			input:    time.Now().Add(0 - Day*3).UnixNano(),
			expected: "3 days",
			err:      nil,
		}, {
			name:     "weeks",
			input:    time.Now().Add(0 - Week*3).UnixNano(),
			expected: "3 weeks",
			err:      nil,
		}, {
			name:     "months",
			input:    time.Now().Add(0 - Month*3).UnixNano(),
			expected: "3 months",
			err:      nil,
		}, {
			name:     "years",
			input:    time.Now().Add(0 - Year*3).UnixNano(),
			expected: "3 years",
			err:      nil,
		}, {
			name:     "invalid",
			input:    time.Now().Add(Second).UnixNano(),
			expected: "",
			err:      errInvalid,
		},
	} {
		output, err := FormatTimeInterval(tc.input)
		assert.Equal(t, tc.err, err, tc.name)
		assert.Equal(t, tc.expected, output, tc.name)
	}

}

func TestMerge(t *testing.T) {
	type tMerge struct {
		src      interface{}
		dest     interface{}
		expected interface{}
		ok       bool
	}

	type nestS struct {
		Na int
	}

	type simple struct {
		Sa int
		Sb string
		Sc bool
		Sd map[string]string
		Se nestS
	}

	getIntAddr := func(i int) *int {
		return &i
	}

	assert := assert.New(t)
	for _, tm := range []tMerge{
		{
			expected: "merged object can not be nil",
			ok:       false,
		}, {
			src:      nestS{Na: 1},
			dest:     nestS{Na: 2},
			expected: "merged object not pointer",
			ok:       false,
		}, {
			src:      &nestS{Na: 1},
			dest:     &simple{Sa: 2},
			expected: "src and dest object type must same",
			ok:       false,
		}, {
			src:      getIntAddr(1),
			dest:     getIntAddr(2),
			expected: "merged object type should be struct",
			ok:       false,
		}, {
			src:      &nestS{},
			dest:     &nestS{},
			expected: &nestS{},
			ok:       true,
		}, {
			src:      nestS{Na: 1},
			dest:     &nestS{Na: 2},
			expected: &nestS{Na: 1},
			ok:       true,
		}, {
			src:      &simple{Sa: 1, Sc: true, Sd: map[string]string{"go": "gogo"}, Se: nestS{Na: 11}},
			dest:     &simple{Sa: 2, Sb: "world", Sc: false, Sd: map[string]string{"go": "gogo"}, Se: nestS{Na: 22}},
			expected: &simple{Sa: 1, Sb: "world", Sc: true, Sd: map[string]string{"go": "gogo"}, Se: nestS{Na: 11}},
			ok:       true,
		}, {
			src:      &simple{},
			dest:     &simple{Sa: 1, Sb: "hello", Sc: true, Sd: map[string]string{"go": "gogo"}, Se: nestS{Na: 22}},
			expected: &simple{Sa: 1, Sb: "hello", Sc: true, Sd: map[string]string{"go": "gogo"}, Se: nestS{Na: 22}},
			ok:       true,
		},
	} {
		err := Merge(tm.src, tm.dest)
		if tm.ok {
			assert.NoError(err)
			assert.Equal(tm.expected, tm.dest)
		} else {
			errMsg, ok := tm.expected.(string)
			if !ok {
				t.Fatalf("test should fail: %v", tm)
			}
			assert.EqualError(err, errMsg)
		}
	}
}

func TestDeDuplicate(t *testing.T) {
	type args struct {
		input []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "nil test case",
			args: args{
				input: nil,
			},
			want: nil,
		},
		{
			name: "two duplicated case",
			args: args{
				input: []string{"asdfgh", "asdfgh"},
			},
			want: []string{"asdfgh"},
		},
		{
			name: "case with no duplicated",
			args: args{
				input: []string{"asdfgh01", "asdfgh02", "asdfgh03", "asdfgh04"},
			},
			want: []string{"asdfgh01", "asdfgh02", "asdfgh03", "asdfgh04"},
		},
		{
			name: "case with no duplicated",
			args: args{
				input: []string{"asdfgh01", "asdfgh02", "asdfgh01"},
			},
			want: []string{"asdfgh01", "asdfgh02"},
		},
		{
			name: "case with 3 duplicated",
			args: args{
				input: []string{"asdfgh01", "asdfgh01", "asdfgh01"},
			},
			want: []string{"asdfgh01"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeDuplicate(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeDuplicate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombineErrors(t *testing.T) {
	formatErrMsg := func(idx int, err error) (string, error) {
		return "Error: " + err.Error(), nil
	}
	errs := []error{
		goerrors.New("Fetch object error: No such object: alpine"),
		goerrors.New("Template parsing error: Can't evaluate field Name"),
	}
	combinedErr := CombineErrors(errs, formatErrMsg)
	expectedErrMsg := "Error: Fetch object error: No such object: alpine\n" +
		"Error: Template parsing error: Can't evaluate field Name"
	if combinedErr.Error() != expectedErrMsg {
		t.Errorf("get error: expected: \n%s, but was: \n%s", expectedErrMsg, combinedErr)
	}

	formatErrMsg = func(idx int, err error) (string, error) {
		return "", goerrors.New("Error: failed to format error message")
	}
	combinedErr = CombineErrors(errs, formatErrMsg)
	expectedErrMsg = "Combine errors error: Error: failed to format error message"
	if combinedErr.Error() != expectedErrMsg {
		t.Errorf("get error: expected: %s, but was: %s", expectedErrMsg, combinedErr)
	}
}

func TestContains(t *testing.T) {
	type args struct {
		input []interface{}
		value interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{input: []interface{}{1, 2}, value: "1"}, want: false, wantErr: true},
		{name: "test2", args: args{input: []interface{}{"1", "2"}, value: "1"}, want: true, wantErr: false},
		{name: "test3", args: args{input: []interface{}{"1", "2"}, value: "3"}, want: false, wantErr: false},
		{name: "test4", args: args{input: []interface{}{1, 2}, value: 1}, want: true, wantErr: false},
		{name: "test5", args: args{input: []interface{}{1, 2}, value: 3}, want: false, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Contains(tt.args.input, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Contains() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringInSlice(t *testing.T) {
	type args struct {
		str   string
		input []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "TestInSlice", args: args{input: []string{"foo", "bar"}, str: "foo"}, want: true},
		{name: "TestNotInSlice", args: args{input: []string{"goods", "bar"}, str: "foo"}, want: false},
		{name: "TestEmptyStr", args: args{input: []string{"foo", "bar"}, str: ""}, want: false},
		{name: "TestEmptySlice", args: args{input: []string{}, str: "bar"}, want: false},
		{name: "TestAllEmpty", args: args{input: []string{}, str: ""}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringInSlice(tt.args.input, tt.args.str); got != tt.want {
				t.Errorf("StringInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
