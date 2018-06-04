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

func TestVolumeListServerError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.VolumeList(context.Background())
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestVolumeList(t *testing.T) {
	expectedURL := "/volumes"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", req.Method)
		}

		volListResp, err := json.Marshal(types.VolumeListResp{
			Volumes: []*types.VolumeInfo{
				{
					Driver: "local",
					Name:   "volume-1",
				},
				{
					Driver: "local",
					Name:   "volume-2",
				},
				{
					Driver: "local",
					Name:   "volume-3",
					Scope:  "global",
				},
			},
			Warnings: []string{"warning"},
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(volListResp))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	volume, err := client.VolumeList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(volume.Volumes), 3)
}
