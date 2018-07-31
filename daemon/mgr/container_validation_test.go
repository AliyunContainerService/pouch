package mgr

import (
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
