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

func TestVolumeInspectServerError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.VolumeInspect(context.Background(), "volume_id")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestVolumeInspectNoFoundError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusConflict, "Not Found")),
	}
	_, err := client.VolumeInspect(context.Background(), "no volume")
	if err == nil || !strings.Contains(err.Error(), "Not Found") {
		t.Fatalf("expected a Volume Not Found Error, got %v", err)
	}
}

func TestVolumeInspect(t *testing.T) {
	expectedURL := "/volumes/volume_id"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", req.Method)
		}

		volInspectResp, err := json.Marshal(types.VolumeInfo{
			Driver: "local",
			Name:   "volume-1",
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(volInspectResp))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	volume, err := client.VolumeInspect(context.Background(), "volume_id")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, volume.Name, "volume-1")
	assert.Equal(t, volume.Driver, "local")
}
