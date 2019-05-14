package formatter

import (
	"testing"

	"github.com/alibaba/pouch/apis/types"

	"github.com/stretchr/testify/assert"
)

func TestNewContainerContext(t *testing.T) {
	type TestCase struct {
		container   *types.Container
		flagNoTrunc bool
		expected    ContainerContext
	}

	testCases := []TestCase{
		{
			container: &types.Container{
				Command:    "bash",
				Created:    8392772900000,
				HostConfig: &types.HostConfig{Runtime: "runc", PortBindings: types.PortMap{"80/tcp": []types.PortBinding{{HostIP: "127.0.0.1", HostPort: "80"}, {HostIP: "127.0.0.1", HostPort: "88"}}}, NetworkMode: "bridge"},
				ID:         "abcdelj8937",
				Image:      "Image123",
				ImageID:    "234567890",
				Labels:     map[string]string{"a": "b", "c": "d"},
				Names:      []string{"nameA", "nameB"},
				State:      "StateA",
				Status:     "StatusB",
				Mounts: []types.MountPoint{
					{Source: "/root/code", Driver: "local"},
					{Source: "/test"},
				},
				SizeRw:     10,
				SizeRootFs: 100,
			},
			flagNoTrunc: false,
			expected: ContainerContext{
				"Names":        "nameA",
				"ID":           "abcdel",
				"Status":       "StatusB",
				"RunningFor":   "49 years" + " ago",
				"Image":        "Image123",
				"Runtime":      "runc",
				"Command":      "bash",
				"ImageID":      "234567890",
				"Labels":       "a = b;c = d;",
				"Mounts":       "/root/code;/test;",
				"State":        "StateA",
				"Ports":        "80/tcp->127.0.0.1:80;80/tcp->127.0.0.1:88;",
				"Size":         "10B (virtual 100B)",
				"LocalVolumes": "1",
				"Networks":     "bridge",
				"CreatedAt":    "1970-01-01 02:19:52.7729 +0000 UTC",
			},
		},
		{
			container: &types.Container{
				Command:    "shell",
				Created:    997349794700000,
				HostConfig: &types.HostConfig{Runtime: "runv"},
				ID:         "1234567890",
				Image:      "Image456",
				ImageID:    "abcdefg",
				Labels:     map[string]string{},
				Names:      []string{"nameB", "nameA"},
				State:      "StateB",
				Status:     "StatusA",
				Mounts: []types.MountPoint{
					{Source: "/root/code"},
					{Source: "/test"},
				},
			},
			flagNoTrunc: true,
			expected: ContainerContext{
				"Names":        "nameB",
				"ID":           "1234567890",
				"Status":       "StatusA",
				"RunningFor":   "49 years" + " ago",
				"Image":        "Image456",
				"Runtime":      "runv",
				"Command":      "shell",
				"ImageID":      "abcdefg",
				"Labels":       "",
				"Mounts":       "/root/code;/test;",
				"State":        "StateB",
				"Ports":        "",
				"Size":         "0B",
				"LocalVolumes": "0",
				"Networks":     "",
				"CreatedAt":    "1970-01-12 13:02:29.7947 +0000 UTC",
			},
		},
	}
	for _, testCase := range testCases {
		result, _ := NewContainerContext(testCase.container, testCase.flagNoTrunc)
		assert.Equal(t, testCase.expected, result)
	}
}
