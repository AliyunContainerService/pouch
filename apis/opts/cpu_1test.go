package opts

import (
	"testing"
        "fmt"
        "reflect"
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
func TestValidateCPUPeriod(t *testing.T) {
	type args struct {
    		env []string
    	}
    	tests := []struct {
    		name    string
    		args    int64
    		want    error
    		wantErr bool
    	}{
    		{name: "test1", args: 0, want: nil, wantErr: true},
    		{name: "test2", args: 500,want:fmt.Errorf("CPU cfs period 500 cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)"), wantErr: false},
    		{name: "test3", args: 1000, want: nil, wantErr: true},
    	}
    	for _, tt := range tests {
    		t.Run(tt.name, func(t *testing.T) {
                        fmt.Print(tt.args);
    			err := ValidateCPUPeriod(tt.args)
    			if (err != nil) != tt.wantErr {

    				return
    			}
    			if !reflect.DeepEqual(err, tt.want) {
    			}
    		})
    	} 
    
}

