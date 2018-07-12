package opts

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateCPUPeriod(t *testing.T) {
	type args struct {
		period int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{

		{name: "test1", args: args{period: 0}, wantErr: nil},
		{name: "test2", args: args{period: 900}, wantErr: fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", 900)},
		{name: "test3", args: args{period: 1100000}, wantErr: fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", 1100000)},
		{name: "test4", args: args{period: 100000}, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCPUPeriod(tt.args.period)
			assert.Equal(t, tt.wantErr, err)

		})
	}
}

func TestValidateCPUQuota(t *testing.T) {
	type args struct {
		period int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{

		{name: "test1", args: args{period: 0}, wantErr: nil},
		{name: "test2", args: args{period: 900}, wantErr: fmt.Errorf("CPU cfs quota %d cannot be less than 1ms (i.e. 1000)", 900)},
		{name: "test3", args: args{period: 100000}, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCPUQuota(tt.args.period)
			assert.Equal(t, tt.wantErr, err)

		})
	}
}
