package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestVolumeRemoveNotFoundError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusNotFound, "Not Found")),
	}
	err := client.VolumeRemove(context.Background(), "no network")
	if err == nil || !strings.Contains(err.Error(), "Not Found") {
		t.Fatalf("expected a Not Found Error, got %v", err)
	}
}

func TestVolumeRemove(t *testing.T) {
	expectedURL := "/volumes/volume_id"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Method != "DELETE" {
			return nil, fmt.Errorf("expected DELETE method, got %s", req.Method)
		}

		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	err := client.VolumeRemove(context.Background(), "volume_id")
	if err != nil {
		t.Fatal(err)
	}
}
