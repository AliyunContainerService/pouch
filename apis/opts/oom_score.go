package opts

import (
	"fmt"
	"testing"
	"reflect"
)

// ValidateOOMScore validates oom score
func ValidateOOMScore(score int64) error {
	if score < -1000 || score > 1000 {
		return fmt.Errorf("oom-score-adj should be in range [-1000, 1000]")
	}

	return nil
}

func TestValidateOOMScore(t *testing.T) {
	type args struct {
		env []string
	}
	tests := []struct {
		name    string
		args    int64
		wantErr bool
	}{
		{name: "test1", args: -1001, wantErr: true},
		{name: "test2", args: 0,     wantErr: false},
		{name: "test2", args: 1001,  wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateOOMScore(tt.args)
			if (got != nil) {
				t.Errorf("oom-score-adj should be in range [-1000, 1000]")
				return
			}
			if got == nil {
				return
			}
		})
	}
}
