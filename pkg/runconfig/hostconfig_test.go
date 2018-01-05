package runconfig

import (
	"fmt"
	"testing"

	"github.com/alibaba/pouch/apis/types"

	"github.com/stretchr/testify/assert"
)

type tDeviceModeCase struct {
	input    string
	expected bool
}

type tParseDeviceCase struct {
	input    string
	expected *types.DeviceMapping
	err      error
}

func TestParseDevice(t *testing.T) {
	for _, tc := range []tParseDeviceCase{
		{
			input: "/dev/zero:/dev/zero:rwm",
			expected: &types.DeviceMapping{
				PathOnHost:        "/dev/zero",
				PathInContainer:   "/dev/zero",
				CgroupPermissions: "rwm",
			},
			err: nil,
		}, {
			input: "/dev/zero:rwm",
			expected: &types.DeviceMapping{
				PathOnHost:        "/dev/zero",
				PathInContainer:   "rwm",
				CgroupPermissions: "rwm",
			},
			err: nil,
		}, {
			input: "/dev/zero",
			expected: &types.DeviceMapping{
				PathOnHost:        "/dev/zero",
				PathInContainer:   "/dev/zero",
				CgroupPermissions: "rwm",
			},
			err: nil,
		}, {
			input: "/dev/zero:/dev/testzero:rwm",
			expected: &types.DeviceMapping{
				PathOnHost:        "/dev/zero",
				PathInContainer:   "/dev/testzero",
				CgroupPermissions: "rwm",
			},
			err: nil,
		}, {
			input:    "/dev/zero:/dev/testzero:rwm:tooLong",
			expected: nil,
			err:      fmt.Errorf("invalid device specification: /dev/zero:/dev/testzero:rwm:tooLong"),
		},
	} {
		output, err := ParseDevice(tc.input)
		assert.Equal(t, tc.err, err, tc.input)
		assert.Equal(t, tc.expected, output, tc.input)
	}
}

func TestValidDeviceMode(t *testing.T) {
	for _, modeCase := range []tDeviceModeCase{
		{
			input:    "rwm",
			expected: true,
		}, {
			input:    "r",
			expected: true,
		}, {
			input:    "rw",
			expected: true,
		}, {
			input:    "rr",
			expected: false,
		}, {
			input:    "rxm",
			expected: false,
		},
	} {
		isValid := ValidDeviceMode(modeCase.input)
		assert.Equal(t, modeCase.expected, isValid, modeCase.input)
	}
}
