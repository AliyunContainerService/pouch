package httputils

import (
	"net/http"
	"testing"
)

func TestBoolValue(t *testing.T) {
	type args struct {
		r *http.Request
		k string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BoolValue(tt.args.r, tt.args.k); got != tt.want {
				t.Errorf("BoolValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
