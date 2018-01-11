package mgr

import (
	"testing"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/stretchr/testify/assert"
)

type tCase struct {
	name     string
	input    *ContainerMeta
	expected string
	err      error
}

func TestContainerMeta_FormatStatus(t *testing.T) {
	// TODO: add more cases
	for _, tc := range []tCase{
		{
			name: "Created",
			input: &ContainerMeta{
				State: &types.ContainerState{
					Status: types.StatusCreated,
				},
			},
			expected: string(types.StatusCreated),
			err:      nil,
		},
		{
			name: "Stopped",
			input: &ContainerMeta{
				State: &types.ContainerState{
					Status: types.StatusStopped,
				},
			},
			expected: string(types.StatusStopped),
			err:      nil,
		},
		{
			name: "Running",
			input: &ContainerMeta{
				State: &types.ContainerState{
					Status:    types.StatusRunning,
					StartedAt: time.Now().Add(0 - utils.Minute).UTC().Format(utils.TimeLayout),
				},
			},
			expected: "Up 1 minute",
			err:      nil,
		},
		{
			name: "Paused",
			input: &ContainerMeta{
				State: &types.ContainerState{
					Status:    types.StatusPaused,
					StartedAt: time.Now().Add(0 - utils.Minute*2).UTC().Format(utils.TimeLayout),
				},
			},
			expected: "Up 2 minutes(paused)",
			err:      nil,
		},
	} {
		output, err := tc.input.FormatStatus()
		assert.Equal(t, output, tc.expected, tc.name)
		assert.Equal(t, err, tc.err, tc.name)
	}
}
