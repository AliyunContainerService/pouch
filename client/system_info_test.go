package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/alibaba/pouch/apis/types"

	"github.com/stretchr/testify/assert"
)

func TestSystemInfoError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.SystemInfo(context.Background())
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestSystemInfo(t *testing.T) {
	expectedURL := "/info"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		info := types.SystemInfo{
			ContainersRunning: 2,
			ContainersStopped: 3,
			Debug:             true,
			Name:              "my_host",
			PouchRootDir:      "/var/lib/pouch",
		}
		b, err := json.Marshal(info)
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(b))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	info, err := client.SystemInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, info.PouchRootDir, "/var/lib/pouch")
	assert.Equal(t, info.Name, "my_host")
	assert.Equal(t, info.Debug, true)
	assert.Equal(t, info.ContainersStopped, int64(3))
	assert.Equal(t, info.ContainersRunning, int64(2))
}
