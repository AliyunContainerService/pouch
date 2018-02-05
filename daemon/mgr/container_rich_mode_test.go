package mgr

import (
	"strings"
	"testing"

	"github.com/alibaba/pouch/apis/types"

	"github.com/stretchr/testify/assert"
)

func convertEnvArrayToMap(envs []string) map[string]string {
	m := map[string]string{}
	for _, e := range envs {
		kvs := strings.Split(e, "=")
		if len(kvs) == 2 {
			m[kvs[0]] = kvs[1]
		}
	}

	return m
}

func TestRichModeSet(t *testing.T) {
	c := &ContainerMeta{
		Config: &types.ContainerConfig{
			Rich: true,
		},
	}

	envs1 := richContainerModeEnv(c)
	mEnvs1 := convertEnvArrayToMap(envs1)

	assert.Equal(t, "true", mEnvs1["rich_mode"])
	assert.Equal(t, types.ContainerConfigRichModeDumbInit, mEnvs1[richModeLaunchEnv])

	//test not set rich mode
	c = &ContainerMeta{
		Config: &types.ContainerConfig{
			Rich: false,
		},
	}
	envs2 := richContainerModeEnv(c)
	assert.Equal(t, 0, len(envs2))

	//test set rich mode manner, not set rich mode
	c = &ContainerMeta{
		Config: &types.ContainerConfig{
			Rich:     false,
			RichMode: types.ContainerConfigRichModeSystemd,
		},
	}
	envs3 := richContainerModeEnv(c)
	assert.Equal(t, 0, len(envs3))

	//test set rich mode manner
	c = &ContainerMeta{
		Config: &types.ContainerConfig{
			Rich:     true,
			RichMode: types.ContainerConfigRichModeSystemd,
		},
	}
	envs4 := richContainerModeEnv(c)
	mEnvs4 := convertEnvArrayToMap(envs4)
	assert.Equal(t, "true", mEnvs4["rich_mode"])
	assert.Equal(t, types.ContainerConfigRichModeSystemd, mEnvs4[richModeLaunchEnv])

	//test set rich mode by env
	c = &ContainerMeta{
		Config: &types.ContainerConfig{
			Env: []string{
				interRichModeEnv,
			},
		},
	}
	envs5 := richContainerModeEnv(c)
	mEnvs5 := convertEnvArrayToMap(envs5)
	assert.Equal(t, "true", mEnvs5["rich_mode"])
	assert.Equal(t, types.ContainerConfigRichModeSystemd, mEnvs5[richModeLaunchEnv])
}
