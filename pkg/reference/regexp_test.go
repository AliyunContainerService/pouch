package reference

import (
	"reflect"
	"regexp"
	"testing"
)

func Test_expression(t *testing.T) {
	type args struct {
		literal string
	}
	tests := []struct {
		name string
		args args
		want *regexp.Regexp
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expression(tt.args.literal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_concat(t *testing.T) {
	type args struct {
		exp []*regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want *regexp.Regexp
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := concat(tt.args.exp...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("concat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_entire(t *testing.T) {
	type args struct {
		exp []*regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want *regexp.Regexp
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := entire(tt.args.exp...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entire() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_zeroOrMore(t *testing.T) {
	type args struct {
		exp []*regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want *regexp.Regexp
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := zeroOrMore(tt.args.exp...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("zeroOrMore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_oneOrMore(t *testing.T) {
	type args struct {
		exp []*regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want *regexp.Regexp
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := oneOrMore(tt.args.exp...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("oneOrMore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_group(t *testing.T) {
	type args struct {
		exp []*regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want *regexp.Regexp
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := group(tt.args.exp...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("group() = %v, want %v", got, tt.want)
			}
		})
	}
}
