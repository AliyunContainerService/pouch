package utils

import (
	goerrors "errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		Sf []string
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
			expected: &simple{Sa: 1, Sb: "hello", Sc: false, Sd: map[string]string{"go": "gogo"}, Se: nestS{Na: 22}},
			ok:       true,
		}, {
			src:      &simple{Sa: 1, Sc: true, Sd: map[string]string{"go": "gogo"}, Se: nestS{Na: 11}, Sf: []string{"foo"}},
			dest:     &simple{Sa: 2, Sb: "!", Sc: false, Sd: map[string]string{"go": "old"}, Se: nestS{Na: 22}, Sf: []string{"foo"}},
			expected: &simple{Sa: 1, Sb: "!", Sc: true, Sd: map[string]string{"go": "gogo"}, Se: nestS{Na: 11}, Sf: []string{"foo", "foo"}},
			ok:       true,
		}, {
			src:      &simple{Sa: 0, Sc: false, Sd: map[string]string{"go": "gogo"}, Se: nestS{Na: 11}, Sf: []string{"foo"}},
			dest:     &simple{Sa: 2, Sb: "world", Sc: true, Sd: map[string]string{"go": "old"}, Se: nestS{Na: 22}, Sf: []string{"foo"}},
			expected: &simple{Sa: 2, Sb: "world", Sc: false, Sd: map[string]string{"go": "gogo"}, Se: nestS{Na: 11}, Sf: []string{"foo", "foo"}},
			ok:       true,
		}, {
			src:      &simple{Sd: map[string]string{"go": "gogo", "a": "b"}},
			dest:     &simple{Sd: map[string]string{"go": "old"}},
			expected: &simple{Sd: map[string]string{"go": "gogo", "a": "b"}},
			ok:       true,
		}, {
			src:      &simple{Sd: map[string]string{"go": "gogo", "a": "b"}},
			dest:     &simple{},
			expected: &simple{Sd: map[string]string{"go": "gogo", "a": "b"}},
			ok:       true,
		}, {
			// empty map should not overwrite
			src:      &simple{Sd: map[string]string{}},
			dest:     &simple{Sd: map[string]string{"a": "b"}},
			expected: &simple{Sd: map[string]string{"a": "b"}},
			ok:       true,
		}, {
			// empty slice should not overwrite
			src:      &simple{Sf: []string{}},
			dest:     &simple{Sf: []string{"c"}},
			expected: &simple{Sf: []string{"c"}},
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
			name: "case with two duplicated",
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

func TestCheckPidExist(t *testing.T) {
	assert := assert.New(t)

	type tCase struct {
		path     string
		pidexist bool
	}

	// mock pidfiles with a must-exist pid 1 and a must-not-exist pid 1 << 30
	dir, err := ioutil.TempDir("/tmp/", "")
	assert.NoError(err)
	defer os.RemoveAll(dir)
	file1 := filepath.Join(dir, "file1")
	file2 := filepath.Join(dir, "file2")
	err = ioutil.WriteFile(file1, []byte(fmt.Sprintf("%d", 1)), 0644)
	assert.NoError(err)
	err = ioutil.WriteFile(file2, []byte(fmt.Sprintf("%d", 1<<30)), 0644)
	assert.NoError(err)

	for _, t := range []tCase{
		{
			path:     "/foo/bar",
			pidexist: false,
		},
		{
			path:     file1,
			pidexist: true,
		},
		{
			path:     file2,
			pidexist: false,
		},
	} {
		err := checkPidfileStatus(t.path)
		if t.pidexist {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}
	}
}

func TestConvertKVStringsToMap(t *testing.T) {
	type tCases struct {
		input    []string
		expected map[string]string
		hasError bool
	}

	for idx, tc := range []tCases{
		{
			input:    nil,
			expected: map[string]string{},
			hasError: false,
		}, {
			input:    []string{"withoutValue"},
			expected: nil,
			hasError: true,
		}, {
			input: []string{"key=value"},
			expected: map[string]string{
				"key": "value",
			},
			hasError: false,
		}, {
			input: []string{"key=key=value"},
			expected: map[string]string{
				"key": "key=value",
			},
			hasError: false,
		}, {
			input: []string{"test=1", "flag=oops", "test=2"},
			expected: map[string]string{
				"test": "2",
				"flag": "oops",
			},
			hasError: false,
		},
	} {
		got, err := ConvertKVStringsToMap(tc.input)
		if err == nil && tc.hasError {
			t.Fatalf("[%d case] should have error here, but got nothing", idx)
		}
		if err != nil && !tc.hasError {
			t.Fatalf("[%d case] should have no error here, but got error(%v)", idx, err)
		}

		if !reflect.DeepEqual(got, tc.expected) {
			t.Fatalf("[%d case] should have (%v), but got (%v)", idx, tc.expected, got)
		}
	}
}

func TestConvertKVStrToMapWithNoErr(t *testing.T) {
	type args struct {
		values []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "normal case",
			args: args{[]string{"a=b", "c=d"}},
			want: map[string]string{"a": "b", "c": "d"},
		},
		{
			name: "normal case with empty string",
			args: args{[]string{"a=b", ""}},
			want: map[string]string{"a": "b"},
		},
		{
			name: "normal case with duplicated key but with different value",
			args: args{[]string{"a=b", "a==bb"}},
			want: map[string]string{"a": "=bb"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertKVStrToMapWithNoErr(tt.args.values); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertKVStrToMapWithNoErr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertStrToKV(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name: "normal case",
			args: args{
				input: "a=b",
			},
			want:    "a",
			want1:   "b",
			wantErr: false,
		},
		{
			name: "normal case",
			args: args{
				input: "a=b===",
			},
			want:    "a",
			want1:   "b===",
			wantErr: false,
		},
		{
			name: "empty failure case",
			args: args{
				input: "",
			},
			want:    "",
			want1:   "",
			wantErr: true,
		},
		{
			name: "no equal mark failure case",
			args: args{
				input: "asdfghjk",
			},
			want:    "",
			want1:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ConvertStrToKV(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertStrToKV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertStrToKV() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ConvertStrToKV() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestIsFileExist(t *testing.T) {
	assert := assert.New(t)
	tempDir, err := ioutil.TempDir("/tmp", "")
	assert.NoError(err)
	defer os.RemoveAll(tempDir)
	existPath := make([]string, 0)
	for _, v := range []string{
		"a", "b", "c", "d", "e",
	} {
		path := filepath.Join(tempDir, v)
		existPath = append(existPath, path)
		os.Create(path)
	}

	for _, t := range []struct {
		path  string
		exist bool
	}{
		{
			path:  existPath[0],
			exist: true,
		},
		{
			path:  existPath[1],
			exist: true,
		},
		{
			path:  existPath[2],
			exist: true,
		},
		{
			path:  existPath[3],
			exist: true,
		},
		{
			path:  existPath[4],
			exist: true,
		},
		{
			path:  filepath.Join(tempDir, "foo"),
			exist: false,
		},
		{
			path:  filepath.Join(tempDir, "bar"),
			exist: false,
		},
		{
			path:  filepath.Join(tempDir, "foo/bar"),
			exist: false,
		},
	} {
		assert.Equal(IsFileExist(t.path), t.exist)
	}
}

func TestStringSliceEqual(t *testing.T) {
	tests := []struct {
		s1    []string
		s2    []string
		equal bool
	}{
		{nil, nil, true},
		{nil, []string{"a"}, false},
		{[]string{"a"}, []string{"a"}, true},
		{[]string{"a"}, []string{"b", "a"}, false},
		{[]string{"a", "b"}, []string{"b", "a"}, true},
	}

	for _, test := range tests {
		result := StringSliceEqual(test.s1, test.s2)
		if result != test.equal {
			t.Fatalf("StringSliceEqual(%v, %v) expected: %v, but got %v", test.s1, test.s2, test.equal, result)
		}
	}
}

func TestMergeMap(t *testing.T) {
	type Expect struct {
		err   error
		key   string
		value interface{}
	}
	tests := []struct {
		m1     map[string]interface{}
		m2     map[string]interface{}
		expect Expect
	}{
		{nil, nil, Expect{fmt.Errorf("all of maps are nil"), "", nil}},
		{nil, map[string]interface{}{"a": "a"}, Expect{nil, "a", "a"}},
		{map[string]interface{}{"a": "a"}, nil, Expect{nil, "a", "a"}},
		{map[string]interface{}{"a": "a"}, map[string]interface{}{"a": "b"}, Expect{nil, "a", "b"}},
		{map[string]interface{}{"a": "a"}, map[string]interface{}{"b": "b"}, Expect{nil, "b", "b"}},
	}

	for _, test := range tests {
		m3, err := MergeMap(test.m1, test.m2)
		if err != nil {
			if test.expect.err.Error() != err.Error() {
				t.Fatalf("MergeMap(%v, %v) expected: %v, but got %v", test.m1, test.m2, test.expect.err, err)
			}
		} else {
			if m3[test.expect.key] != test.expect.value {
				t.Fatalf("MergeMap(%v, %v) expected: %v, but got %v", test.m1, test.m2, test.expect.value, m3[test.expect.key])
			}
		}
	}
}
