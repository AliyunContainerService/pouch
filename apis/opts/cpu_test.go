package opts

import (
	"testing"
        "fmt"
)

func ValidateCPUPeriod(period int64) error {
	if period == 0 {
		return nil
	}
	if period < 1000 || period > 1000000 {
		return fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", period)
	}
	return nil
}

func ValidateCPUQuota(quota int64) error {
	if quota == 0 {
		return nil
	}
	if quota < 1000 {
		return fmt.Errorf("CPU cfs quota %d cannot be less than 1ms (i.e. 1000)", quota)
	}
	return nil
}
func TestValidateCPUPeriod(t *testing.T) {
	type args struct {
    		env []string
    	}
    	tests := []struct {
    		name    string
    		args    int64
    		wantErr bool
    	}{
    		{name: "test1", args: 0,  wantErr: true},
    		{name: "test2", args: 500, wantErr: false},
    		{name: "test3", args: 1000,  wantErr: true},
    	}
    	for _, tt := range tests {
    		t.Run(tt.name, func(t *testing.T) {
    			err := ValidateCPUPeriod(tt.args)
    			if (err != nil) != tt.wantErr {
					t.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", tt.args)
    				return
    			}
    			if (err != nil) == tt.wantErr {
    			}
    		})
    	}
}

func TestValidateCPUQuota(t *testing.T) {
	type args struct {
    		env []string
    	}
    	tests := []struct {
    		name    string
    		args    int64
    		wantErr bool
    	}{
    		{name: "test1", args: 0,  wantErr: true},
    		{name: "test2", args: 999, wantErr: false},
    		{name: "test3", args: 1001,  wantErr: true},
    	}
    	for _, tt := range tests {
    		t.Run(tt.name, func(t *testing.T) {
    			err := ValidateCPUQuota(tt.args)
    			if (err != nil) != tt.wantErr {
                    //t.Errorf("CPU cfs quota %d cannot be less than 1ms (i.e. 1000)", tt.args)
    				return
    			}
    			if (err != nil) == tt.wantErr {
					return
    			}
    		})
    	} 
    
}

