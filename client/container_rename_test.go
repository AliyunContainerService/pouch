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
)

func TestContainerRenameError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	err := client.ContainerRename(context.Background(), "nothing", "containerRename")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerRename(t *testing.T) {
	expectedURL := "/containers/container_id/rename"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Header.Get("Content-Type") == "application/json" {
			var renameConfig interface{}
			if err := json.NewDecoder(req.Body).Decode(&renameConfig); err != nil {
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

	if err := client.ContainerRename(context.Background(), "container_id", "containerRename"); err != nil {
		t.Fatal(err)
	}
}
