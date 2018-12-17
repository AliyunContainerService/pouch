package httputils

import (
	"net/http"
	"strings"
	"testing"
)

func TestBoolValue(t *testing.T) {
	tests := []struct {
		name    string
		testStr string
		want    bool
	}{
		{"test normal string", "normal string", true},
		{"test ''", "", false},
		{"test 0", "0", false},
		{"test no", "no", false},
		{"test NO", "NO", false},
		{"test false", "false", false},
		{"test FALSE", "FALSE", false},
		{"test none", "none", false},
		{"test NONE", "NONE", false},
		// test trim
		{"test '  '", "  ", false},
		{"test trim 0", "  0  ", false},
		{"test trim no", "  no  ", false},
		{"test trim NO", "  NO ", false},
		{"test trim false", " false ", false},
		{"test trim FALSE", "  FALSE ", false},
		{"test trim none", "  none ", false},
		{"test trim NONE", "  NONE  ", false},
	}
	boolName := "name"
	for _, tt := range tests {
		request, _ := http.NewRequest(
			"POST",
			"http://test",
			strings.NewReader(boolName+"="+tt.testStr),
		)

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		t.Run(tt.name, func(t *testing.T) {
			if got := BoolValue(request, boolName); got != tt.want {
				t.Errorf("%s, BoolValue() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
