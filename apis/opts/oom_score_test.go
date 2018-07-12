package opts

import (
	"testing"
)

func TestValidateOOMScore(t *testing.T) {
	cases := []struct {
		name  string
		score int64
		want  bool
	}{
		{"test1", 0, false},
		{"test2", 10, false},
		{"test3", 100, false},
		{"test4", 1000, false},
		{"test5", -10, false},
		{"test6", -100, false},
		{"test7", -1000, false},
		{"test8", 2000, true},
		{"test9", 5000, true},
		{"test10", -2000, true},
		{"test11", -5000, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ValidateOOMScore(c.score)
			if got != nil && (got != nil) != c.want {
				t.Fatal("TestValidateOOMScore failed")
			}
		})
	}
}
