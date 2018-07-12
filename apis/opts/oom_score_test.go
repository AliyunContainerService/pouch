package opts

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateOOMScore(t *testing.T) {
	type args struct {
		score int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{score: -1001}, wantErr: fmt.Errorf("oom-score-adj should be in range [-1000, 1000]")},
		{name: "test2", args: args{score: 0}, wantErr: nil},
		{name: "test3", args: args{score: 11000}, wantErr: fmt.Errorf("oom-score-adj should be in range [-1000, 1000]")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOOMScore(tt.args.score)
			assert.Equal(t, tt.wantErr, err)

		})
	}
}
