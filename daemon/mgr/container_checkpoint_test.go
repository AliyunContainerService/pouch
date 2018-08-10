package mgr

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/alibaba/pouch/apis/types"

	"github.com/stretchr/testify/assert"
)

func TestRWCheckpointConfig(t *testing.T) {
	assert := assert.New(t)
	tmpDir, err := ioutil.TempDir("", "checkpoint-test")
	assert.NoError(err)
	defer os.RemoveAll(tmpDir)

	for _, t := range []struct {
		path       string
		name       string
		checkpoint string
	}{
		{
			path:       filepath.Join(tmpDir, "1"),
			name:       "n1",
			checkpoint: "c1",
		},
		{
			path:       filepath.Join(tmpDir, "2"),
			name:       "",
			checkpoint: "",
		},
		{
			path:       filepath.Join(tmpDir, "3"),
			name:       "foo",
			checkpoint: "bar",
		},
	} {
		assert.NoError(writeCheckpointConfig(t.path, t.name, t.checkpoint))
		c, err := readCheckpointConfig(t.path)
		assert.NoError(err)
		assert.Equal(c, &types.Checkpoint{
			ContainerID:    t.name,
			CheckpointName: t.checkpoint,
		})
	}
}
