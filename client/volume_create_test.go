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

func TestVolumeCreateServerError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.VolumeCreate(context.Background(), &types.VolumeCreateConfig{})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestVolumeCreateDupError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusConflict, "Volume already exists")),
	}
	_, err := client.VolumeCreate(context.Background(), &types.VolumeCreateConfig{})
	if err == nil || !strings.Contains(err.Error(), "Volume already exists") {
		t.Fatalf("expected a Volume Already Exists Error, got %v", err)
	}
}

func TestVolumeCreate(t *testing.T) {
	expectedURL := "/volumes/create"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Header.Get("Content-Type") == "application/json" {
			var createConfig interface{}
			if err := json.NewDecoder(req.Body).Decode(&createConfig); err != nil {
				return nil, fmt.Errorf("failed to parse json: %v", err)
			}
		}

		volCreateResp, err := json.Marshal(types.VolumeInfo{
			Name:   "volume-1",
			Driver: "local",
		})
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(volCreateResp))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	volume, err := client.VolumeCreate(context.Background(), &types.VolumeCreateConfig{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, volume.Name, "volume-1")
	assert.Equal(t, volume.Driver, "local")

}
