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
)

func TestContainerUpdateError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	err := client.ContainerUpdate(context.Background(), "nothing", &types.UpdateConfig{})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerUpdate(t *testing.T) {
	expectedURL := "/containers/container_id/update"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Header.Get("Content-Type") == "application/json" {
			var updateConfig interface{}
			if err := json.NewDecoder(req.Body).Decode(&updateConfig); err != nil {
				return nil, fmt.Errorf("failed to parse json: %v", err)
			}
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})
	client := &APIClient{
		HTTPCli: httpClient,
	}
	if err := client.ContainerUpdate(context.Background(), "container_id", &types.UpdateConfig{}); err != nil {
		t.Fatal(err)
	}
}
