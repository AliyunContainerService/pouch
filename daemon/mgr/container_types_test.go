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
	input    *Container
	expected string
	err      error
}

func TestContainer_FormatStatus(t *testing.T) {
	// TODO: add more cases
	for _, tc := range []tCase{
		{
			name: "Created",
			input: &Container{
				State: &types.ContainerState{
					Status: types.StatusCreated,
				},
			},
			expected: string(types.StatusCreated),
			err:      nil,
		},
		{
			name: "Exited",
			input: &Container{
				State: &types.ContainerState{
					Status:     types.StatusExited,
					FinishedAt: time.Now().Add(0 - utils.Hour).UTC().Format(utils.TimeLayout),
					ExitCode:   0,
				},
			},
			expected: "Exited (0) 1 hour",
			err:      nil,
		},
		{
			name: "Stopped",
			input: &Container{
				State: &types.ContainerState{
					Status:     types.StatusStopped,
					FinishedAt: time.Now().Add(0 - utils.Minute).UTC().Format(utils.TimeLayout),
					ExitCode:   1,
				},
			},
			expected: "Stopped (1) 1 minute",
			err:      nil,
		},
		{
			name: "Running",
			input: &Container{
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
			input: &Container{
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
