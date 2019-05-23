package mgr

import (
	"fmt"
	"testing"

	"github.com/alibaba/pouch/apis/types"

	"github.com/stretchr/testify/assert"
)

func TestValidateNilNvidiaConfig(t *testing.T) {
	r := types.Resources{
		NvidiaConfig: nil,
	}
	assert.NoError(t, validateNvidiaConfig(&r))
}

func TestValidateNvidiaDevice(t *testing.T) {
	type tCase struct {
		r           types.Resources
		errExpected error
	}

	for _, tc := range []tCase{
		{
			r: types.Resources{
				NvidiaConfig: &types.NvidiaConfig{
					NvidiaVisibleDevices: "all",
				}},
			errExpected: nil,
		},
		{
			r: types.Resources{
				NvidiaConfig: &types.NvidiaConfig{
					NvidiaVisibleDevices: "none",
				},
			},
			errExpected: nil,
		},
		{
			r: types.Resources{
				NvidiaConfig: &types.NvidiaConfig{
					NvidiaVisibleDevices: "void",
				},
			},
			errExpected: nil,
		},
		{
			r: types.Resources{
				NvidiaConfig: &types.NvidiaConfig{
					NvidiaVisibleDevices: "",
				},
			},
			errExpected: nil,
		},
		{
			r: types.Resources{
				NvidiaConfig: &types.NvidiaConfig{
					NvidiaVisibleDevices: "9",
				},
			},
			errExpected: errInvalidDevice,
		},
	} {
		err := validateNvidiaDevice(&tc.r)
		assert.Equal(t, tc.errExpected, err)
	}
}

func TestValidateNvidiaDriver(t *testing.T) {
	type tCase struct {
		r           types.Resources
		errExpected error
	}

	for _, tc := range []tCase{
		{
			r: types.Resources{
				NvidiaConfig: &types.NvidiaConfig{
					NvidiaDriverCapabilities: "all",
				}},
			errExpected: nil,
		},
		{
			r: types.Resources{
				NvidiaConfig: &types.NvidiaConfig{
					NvidiaDriverCapabilities: "",
				},
			},
			errExpected: nil,
		},
		{
			r: types.Resources{
				NvidiaConfig: &types.NvidiaConfig{
					NvidiaDriverCapabilities: "compute,compat32,graphics,utility,video,display",
				},
			},
			errExpected: nil,
		},
		{
			r: types.Resources{
				NvidiaConfig: &types.NvidiaConfig{
					NvidiaDriverCapabilities: "ErrorError",
				},
			},
			errExpected: errInvalidDriver,
		},
	} {
		err := validateNvidiaDriver(&tc.r)
		assert.Equal(t, tc.errExpected, err)
	}
}

func TestValidateResource(t *testing.T) {

	type tCase struct {
		r                types.Resources
		update           bool
		warningsExpected []string
		errExpected      error
	}

	for _, tc := range []tCase{
		{
			r: types.Resources{
				MemoryReservation: 8388608, //8m
			},
			warningsExpected: []string{},
			errExpected:      nil,
		},
		{
			r: types.Resources{
				MemoryReservation: 2097152, //2m
			},
			warningsExpected: []string{},
			errExpected:      fmt.Errorf("Minimal memory reservation should greater than 4M"),
		},
		{
			r: types.Resources{
				Memory:            8388608,
				MemoryReservation: 10485760,
			},
			warningsExpected: []string{},
			errExpected:      fmt.Errorf("Minimum memory limit should be larger than memory reservation limit"),
		},
		{
			r: types.Resources{
				Memory:            8388608,
				MemoryReservation: 10485760,
			},
			warningsExpected: []string{},
			errExpected:      fmt.Errorf("Minimum memory limit should be larger than memory reservation limit"),
		},
		{
			r: types.Resources{
				MemorySwap: 8388608,
				Memory:     10485760,
			},
			warningsExpected: []string{},
			errExpected:      fmt.Errorf("Minimum memoryswap limit should be larger than memory limit"),
		},
		{
			r: types.Resources{
				MemorySwap: 8388608,
				Memory:     0,
			},
			warningsExpected: []string{},
			errExpected:      fmt.Errorf("You should always set the Memory limit when using Memoryswap limit"),
		},
		{
			r: types.Resources{
				Memory: 2097152,
			},
			warningsExpected: []string{},
			errExpected:      fmt.Errorf("Minimal memory should greater than 4M"),
		},
	} {
		warnings, err := validateResource(&tc.r, tc.update)
		assert.Equal(t, tc.warningsExpected, warnings)
		assert.Equal(t, tc.errExpected, err)
	}
}
