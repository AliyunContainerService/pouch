package mgr

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/opencontainers/image-spec/specs-go/v1"

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

func TestMerge(t *testing.T) {
	assert := assert.New(t)

	type TPort struct {
		p int
	}

	type TVolume struct {
		v int
	}

	for idx, tc := range []struct {
		c        *Container
		image    v1.ImageConfig
		expected *types.ContainerConfig
	}{
		{
			// test merge from image when both config is not empty
			c: &Container{
				Config: &types.ContainerConfig{
					Cmd:        []string{"a"},
					Entrypoint: []string{"b"},
					Env:        []string{"e1=e1", "e2=e2"},
					User:       "user1",
					StopSignal: "1",
					Labels: map[string]string{
						"l1": "l2",
						"l3": "l4",
					},
					ExposedPorts: map[string]interface{}{
						"e1": interface{}(TPort{
							p: 1,
						}),
						"e2": interface{}(TPort{
							p: 2,
						}),
					},
					Volumes: map[string]interface{}{
						"v1": interface{}(TVolume{
							v: 1,
						}),
						"v2": interface{}(TVolume{
							v: 2,
						}),
					},
				},
			},
			image: v1.ImageConfig{
				Cmd:        []string{"ia"},
				Entrypoint: []string{"ib"},
				Env:        []string{"e1=e1", "ie2=ie2"},
				User:       "iuser1",
				StopSignal: "2",
				Labels: map[string]string{
					"il1": "l2",
					"il3": "l4",
				},
				ExposedPorts: map[string]struct{}{
					"ie1": {},
					"e2":  {},
				},
				Volumes: map[string]struct{}{
					"iv1": {},
					"iv2": {},
				},
			},
			expected: &types.ContainerConfig{
				Cmd:        []string{"a"},
				Entrypoint: []string{"b"},
				Env:        []string{"e1=e1", "e2=e2", "ie2=ie2"},
				User:       "user1",
				StopSignal: "1",
				Labels: map[string]string{
					"l3":  "l4",
					"il1": "l2",
					"l1":  "l2",
					"il3": "l4",
				},
				ExposedPorts: map[string]interface{}{
					"e1": interface{}(TPort{
						p: 1,
					}),
					"e2": interface{}(TPort{
						p: 2,
					}),
					"ie1": struct{}{},
				},
				Volumes: map[string]interface{}{
					"v1": interface{}(TVolume{
						v: 1,
					}),
					"v2": interface{}(TVolume{
						v: 2,
					}),
					"iv1": struct{}{},
					"iv2": struct{}{},
				},
			},
		},
		{
			// test merge image config, when image config is empty
			c: &Container{
				Config: &types.ContainerConfig{
					Cmd:        []string{"a"},
					Entrypoint: []string{"b"},
					Env:        []string{"e1=e1", "e2=e2"},
					User:       "user1",
					StopSignal: "1",
					Labels: map[string]string{
						"l1": "l2",
						"l3": "l4",
					},
				},
			},
			image: v1.ImageConfig{},
			expected: &types.ContainerConfig{
				Cmd:        []string{"a"},
				Entrypoint: []string{"b"},
				Env:        []string{"e1=e1", "e2=e2"},
				User:       "user1",
				StopSignal: "1",
				Labels: map[string]string{
					"l1": "l2",
					"l3": "l4",
				},
			},
		},
		{
			// test merge image config, when container config is empty
			c: &Container{
				Config: &types.ContainerConfig{},
			},
			image: v1.ImageConfig{
				Cmd:        []string{"ia"},
				Entrypoint: []string{"ib"},
				Env:        []string{"ie1=ie1", "ie2=ie2"},
				User:       "iuser1",
				StopSignal: "1",
				Labels: map[string]string{
					"il1": "l2",
					"il3": "l4",
				},
			},
			expected: &types.ContainerConfig{
				Cmd:        []string{"ia"},
				Entrypoint: []string{"ib"},
				Env:        []string{"ie1=ie1", "ie2=ie2"},
				User:       "iuser1",
				StopSignal: "1",
				Labels: map[string]string{
					"il1": "l2",
					"il3": "l4",
				},
			},
		},
	} {
		err := tc.c.merge(func() (v1.ImageConfig, error) {
			return tc.image, nil
		})
		assert.NoError(err)

		// sort slice
		sort.Strings(tc.c.Config.Env)
		sort.Strings(tc.expected.Env)

		ret := reflect.DeepEqual(tc.c.Config, tc.expected)
		assert.Equal(true, ret, fmt.Sprintf("test %d fails\n %+v should equal with %+v\n", idx, tc.c.Config, tc.expected))
	}
}
