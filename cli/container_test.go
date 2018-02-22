package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/stretchr/testify/assert"
)

func TestParseRestartPolicy(t *testing.T) {
	type TestCase struct {
		input         string
		expectedName  string
		expectedCount int64
		err           error
	}

	cases := []TestCase{
		{
			input:         "always",
			expectedName:  "always",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "no",
			expectedName:  "no",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "unless-stopped",
			expectedName:  "unless-stopped",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "on-failure:1",
			expectedName:  "on-failure",
			expectedCount: 1,
			err:           nil,
		},
		{
			input:         "on-failure",
			expectedName:  "on-failure",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "on-failure:1:2",
			expectedName:  "on-failure",
			expectedCount: 0,
			err:           fmt.Errorf("invalid restart policy: %s", "on-failure:1:2"),
		},
	}

	for _, cs := range cases {
		policy, err := parseRestartPolicy(cs.input)
		assert.Equal(t, cs.err, err)
		if err == nil {
			assert.Equal(t, cs.expectedName, policy.Name)
			assert.Equal(t, cs.expectedCount, policy.MaximumRetryCount)
		}
	}
}

func TestParseLabels(t *testing.T) {
	type result struct {
		labels map[string]string
		err    error
	}
	type TestCase struct {
		input    []string
		expected result
	}

	testCases := []TestCase{
		{
			input: []string{"a=b"},
			expected: result{
				labels: map[string]string{
					"a": "b",
				},
				err: nil,
			},
		},
		{
			input: []string{"a=b", "a=b"},
			expected: result{
				labels: map[string]string{
					"a": "b",
				},
				err: nil,
			},
		},
		{
			// FIXME: this case should throw error
			input: []string{"a=b", "a=bb"},
			expected: result{
				labels: map[string]string{
					"a": "bb",
				},
				err: nil,
			},
		},
		{
			input: []string{"ThisIsALableWithoutEqualMark"},
			expected: result{
				labels: nil,
				err:    fmt.Errorf("invalid label ThisIsALableWithoutEqualMark: label must be in format of key=value"),
			},
		},
	}

	for _, testCase := range testCases {
		labels, err := parseLabels(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.labels, labels)
	}
}

func TestParseMemory(t *testing.T) {
	type result struct {
		memory int64
		err    error
	}
	type TestCase struct {
		input    string
		expected result
	}

	testCases := []TestCase{
		{
			input: "",
			expected: result{
				memory: 0,
				err:    nil,
			},
		},
		{
			input: "0",
			expected: result{
				memory: 0,
				err:    nil,
			},
		},
		{
			input: "100m",
			expected: result{
				memory: 104857600,
				err:    nil,
			},
		},
		{
			input: "10asdfg",
			expected: result{
				memory: 0,
				err:    fmt.Errorf("invalid size: '%s'", "10asdfg"),
			},
		},
	}

	for _, testCase := range testCases {
		memory, err := parseMemory(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.memory, memory)
	}
}

func TestParseMemorySwap(t *testing.T) {
	type result struct {
		memorySwap int64
		err        error
	}
	type TestCase struct {
		input    string
		expected result
	}

	testCases := []TestCase{
		{
			input: "",
			expected: result{
				memorySwap: 0,
				err:        nil,
			},
		},
		{
			input: "-1",
			expected: result{
				memorySwap: -1,
				err:        nil,
			},
		},
		{
			input: "100m",
			expected: result{
				memorySwap: 104857600,
				err:        nil,
			},
		},
		{
			input: "10asdfg",
			expected: result{
				memorySwap: 0,
				err:        fmt.Errorf("invalid size: '%s'", "10asdfg"),
			},
		},
	}

	for _, testCase := range testCases {
		memorySwap, err := parseMemorySwap(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.memorySwap, memorySwap)
	}
}

func TestValidateMemorySwappiness(t *testing.T) {
	type TestCase struct {
		input    int64
		expected error
	}

	testCases := []TestCase{
		{
			input:    -1,
			expected: nil,
		},
		{
			input:    0,
			expected: nil,
		},
		{
			input:    100,
			expected: nil,
		},
		{
			input:    38,
			expected: nil,
		},
		{
			input:    -5,
			expected: fmt.Errorf("invalid memory swappiness: %d (its range is -1 or 0-100)", -5),
		},
		{
			input:    200,
			expected: fmt.Errorf("invalid memory swappiness: %d (its range is -1 or 0-100)", 200),
		},
	}

	for _, testCase := range testCases {
		err := validateMemorySwappiness(testCase.input)
		assert.Equal(t, testCase.expected, err)
	}
}

func Test_parseDeviceMapping(t *testing.T) {
	type args struct {
		device string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.DeviceMapping
		wantErr bool
	}{
		{
			name: "deviceMapping1",
			args: args{
				device: "/dev/deviceName:/dev/deviceName:mrw",
			},
			want: &types.DeviceMapping{
				PathOnHost:        "/dev/deviceName",
				PathInContainer:   "/dev/deviceName",
				CgroupPermissions: "mrw",
			},

			wantErr: false,
		},
		{
			name: "deviceMapping2",
			args: args{
				device: "/dev/deviceName:",
			},
			want: &types.DeviceMapping{
				PathOnHost:        "/dev/deviceName",
				PathInContainer:   "/dev/deviceName",
				CgroupPermissions: "rwm",
			},
			wantErr: false,
		},
		{
			name: "deviceMappingWrong1",
			args: args{
				device: "/dev/deviceName:/dev/deviceName:rrw",
			},
			wantErr: true,
		},
		{
			name: "deviceMappingWrong2",
			args: args{
				device: "/dev/deviceName:/dev/deviceName:arw",
			},
			wantErr: true,
		},
		{
			name: "deviceMappingWrong3",
			args: args{
				device: "/dev/deviceName:/dev/deviceName:mrw:mrw",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDevice(tt.args.device)
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

func Test_parseSysctls(t *testing.T) {
	type result struct {
		sysctls map[string]string
		err     error
	}
	type TestCases struct {
		input  []string
		expect result
	}

	testCases := []TestCases{
		{
			input: []string{"a=b"},
			expect: result{
				sysctls: map[string]string{"a": "b"},
				err:     nil,
			},
		},
		{
			input: []string{"ab"},
			expect: result{
				sysctls: nil,
				err:     fmt.Errorf("invalid sysctl %s: sysctl must be in format of key=value", "ab"),
			},
		},
	}

	for _, testCase := range testCases {
		sysctl, err := parseSysctls(testCase.input)
		assert.Equal(t, testCase.expect.sysctls, sysctl)
		assert.Equal(t, testCase.expect.err, err)
	}
}

func Test_parseNetwork(t *testing.T) {
	type net struct {
		name      string
		parameter string
		mode      string
	}
	type result struct {
		network net
		err     error
	}
	type TestCases struct {
		input  string
		expect result
	}

	testCases := []TestCases{
		{
			input: "",
			expect: result{
				err:     fmt.Errorf("invalid network: cannot be empty"),
				network: net{name: "", parameter: "", mode: ""},
			},
		},
		{
			input: "121.0.0.1",
			expect: result{
				err:     nil,
				network: net{name: "", parameter: "121.0.0.1", mode: ""},
			},
		},
		{
			input: "myHost",
			expect: result{
				err:     nil,
				network: net{name: "myHost", parameter: "", mode: ""},
			},
		},
		{
			input: "myHost:121.0.0.1",
			expect: result{
				err:     nil,
				network: net{name: "myHost", parameter: "121.0.0.1", mode: ""},
			},
		},
		{
			input: "container:9ca6ac",
			expect: result{
				err:     nil,
				network: net{name: "container", parameter: "9ca6ac", mode: ""},
			},
		},
		{
			input: "bridge:121.0.0.1:mode",
			expect: result{
				err:     nil,
				network: net{name: "bridge", parameter: "121.0.0.1", mode: "mode"},
			},
		},
		{
			input: "bridge:mode",
			expect: result{
				err:     nil,
				network: net{name: "bridge", parameter: "", mode: "mode"},
			},
		},
	}

	for _, testCase := range testCases {
		name, parameter, mode, error := parseNetwork(testCase.input)
		assert.Equal(t, testCase.expect.err, error)
		assert.Equal(t, testCase.expect.network.name, name)
		assert.Equal(t, testCase.expect.network.parameter, parameter)
		assert.Equal(t, testCase.expect.network.mode, mode)
	}
}

func Test_parseIntelRdt(t *testing.T) {
	type args struct {
		intelRdtL3Cbm string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIntelRdt(tt.args.intelRdtL3Cbm)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIntelRdt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseIntelRdt() = %v, want %v", got, tt.want)
			}
		})
	}
}
