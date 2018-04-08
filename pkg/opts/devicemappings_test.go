package opts

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/apis/types"

	"github.com/stretchr/testify/assert"
)

func TestParseDeviceMapping(t *testing.T) {
	type args struct {
		device []string
	}
	tests := []struct {
		name    string
		args    args
		want    []*types.DeviceMapping
		wantErr bool
	}{
		{
			name: "deviceMapping1",
			args: args{
				device: []string{"/dev/deviceName:/dev/deviceName:mrw"},
			},
			want: []*types.DeviceMapping{{
				PathOnHost:        "/dev/deviceName",
				PathInContainer:   "/dev/deviceName",
				CgroupPermissions: "mrw",
			}},

			wantErr: false,
		},
		{
			name: "deviceMapping2",
			args: args{
				device: []string{"/dev/deviceName:"},
			},
			want: []*types.DeviceMapping{{
				PathOnHost:        "/dev/deviceName",
				PathInContainer:   "/dev/deviceName",
				CgroupPermissions: "rwm",
			}},
			wantErr: false,
		},
		{
			name: "deviceMappingWrong1",
			args: args{
				device: []string{"/dev/deviceName:/dev/deviceName:rrw"},
			},
			wantErr: true,
		},
		{
			name: "deviceMappingWrong2",
			args: args{
				device: []string{"/dev/deviceName:/dev/deviceName:arw"},
			},
			wantErr: true,
		},
		{
			name: "deviceMappingWrong3",
			args: args{
				device: []string{"/dev/deviceName:/dev/deviceName:mrw:mrw"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDeviceMappings(tt.args.device)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDeviceMappings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDeviceMappings() = %v, want %v", got, tt.want)
			}
		})
	}
}

type tDeviceModeCase struct {
	input    string
	expected bool
}

type tParseDeviceCase struct {
	input    string
	expected *types.DeviceMapping
	err      error
}

func Test_parseDevice(t *testing.T) {
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
		output, err := parseDevice(tc.input)
		assert.Equal(t, tc.err, err, tc.input)
		assert.Equal(t, tc.expected, output, tc.input)
	}
}

func TestValidateDeviceMode(t *testing.T) {
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
		isValid := ValidateDeviceMode(modeCase.input)
		assert.Equal(t, modeCase.expected, isValid, modeCase.input)
	}
}
