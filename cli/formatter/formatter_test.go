package formatter

import (
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/stretchr/testify/assert"
)

func TestPreFormat(t *testing.T) {
	type TestCase struct {
		input  string
		output string
	}

	testCases := []TestCase{
		{
			input:  "table {{.ID}}\t{{.Status}}\t{{.CreatedAt}}\t{{.Name}}\t{{.Image}}\t{{.State}}",
			output: "{{.ID}}\t{{.Status}}\t{{.CreatedAt}}\t{{.Name}}\t{{.Image}}\t{{.State}}\n",
		},
		{
			input:  "table {{.ID}}\\t{{.Status}}\\t{{.CreatedAt}}\\t{{.Name}}\\t{{.Image}}\\t{{.State}}",
			output: "{{.ID}}\t{{.Status}}\t{{.CreatedAt}}\t{{.Name}}\t{{.Image}}\t{{.State}}\n",
		},
		{
			input:  "table {{.ID}}\t{{.Status}}\t{{.CreatedAt}}\\t{{.Name}}\t{{.Image}}\t{{.State}}\n",
			output: "{{.ID}}\t{{.Status}}\t{{.CreatedAt}}\t{{.Name}}\t{{.Image}}\t{{.State}}\n",
		},
		{
			input:  "raw",
			output: "Name:{{.Names}}\nID:{{.ID}}\nStatus:{{.Status}}\nCreated:{{.RunningFor}}\nImage:{{.Image}}\nRuntime:{{.Runtime}}\n\n",
		},
	}
	for _, testCase := range testCases {
		result := PreFormat(testCase.input)
		assert.Equal(t, testCase.output, result)
	}
}

func TestLabelToString(t *testing.T) {
	type TestCase struct {
		input  map[string]string
		output string
	}
	testCases := []TestCase{
		{
			input: map[string]string{
				"a": "b",
				"c": "d",
			},
			output: "a = b;c = d;",
		},
		{
			input: map[string]string{
				"a": "b",
			},
			output: "a = b;",
		},
		{
			input:  map[string]string{},
			output: "",
		},
	}
	for _, testCase := range testCases {
		result := LabelsToString(testCase.input)
		assert.Equal(t, testCase.output, result)
	}
}

func TestMountPointToString(t *testing.T) {
	type TestCase struct {
		input  []types.MountPoint
		output string
	}
	testCases := []TestCase{
		{
			input: []types.MountPoint{
				{Source: "/root/code"},
				{Source: "/test"},
			},
			output: "/root/code;/test;",
		},
		{
			input: []types.MountPoint{
				{Source: "/root/code"},
			},
			output: "/root/code;",
		},
	}
	for _, testCase := range testCases {
		result := MountPointToString(testCase.input)
		assert.Equal(t, testCase.output, result)
	}
}

func TestLocalVolumes(t *testing.T) {
	type TestCase struct {
		input  []types.MountPoint
		output string
	}
	testCases := []TestCase{
		{
			input: []types.MountPoint{
				{Source: "/root/code", Driver: "local"},
				{Source: "/test", Driver: "local"},
			},
			output: "2",
		},
		{
			input: []types.MountPoint{
				{Source: "/root/code"},
			},
			output: "0",
		},
	}
	for _, testCase := range testCases {
		result := LocalVolumes(testCase.input)
		assert.Equal(t, testCase.output, result)
	}
}

func TestSizeToString(t *testing.T) {
	type TestCase struct {
		input  []int64
		output string
	}
	testCases := []TestCase{
		{
			input:  []int64{45, 56},
			output: "45B (virtual 56B)",
		},
		{
			input:  []int64{65526, 409626},
			output: "65.5kB (virtual 410kB)",
		},
	}
	for _, testCase := range testCases {
		result := SizeToString(testCase.input[0], testCase.input[1])
		assert.Equal(t, testCase.output, result)
	}
}

func TestPortBindingsToString(t *testing.T) {
	type TestCase struct {
		input  types.PortMap
		output string
	}
	testCases := []TestCase{
		{
			input: map[string][]types.PortBinding{
				"80/tcp": {
					{
						HostIP:   "127.0.0.1",
						HostPort: "80",
					},
					{
						HostIP:   "127.0.0.1",
						HostPort: "88",
					},
				},
			},
			output: "80/tcp->127.0.0.1:80;80/tcp->127.0.0.1:88;",
		},
		{
			input: map[string][]types.PortBinding{
				"80/udp": {
					{
						HostIP:   "192.168.1.1",
						HostPort: "65535",
					},
					{
						HostIP:   "192.168.1.3",
						HostPort: "65536",
					},
				},
			},
			output: "80/udp->192.168.1.1:65535;80/udp->192.168.1.3:65536;",
		},
	}
	for _, testCase := range testCases {
		result := PortBindingsToString(testCase.input)
		assert.Equal(t, testCase.output, result)
	}
}
