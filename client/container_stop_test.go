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

func TestContainerStopError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	err := client.ContainerStop(context.Background(), "nothing", "10")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerStop(t *testing.T) {
	expectedURL := "/containers/container_id/stop"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		timeout := req.URL.Query().Get("t")
		if timeout != "10" {
			return nil, fmt.Errorf("timeout not set in URL   properly. Expected '10', got %s", timeout)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})
	client := &APIClient{
		HTTPCli: httpClient,
	}
	err := client.ContainerStop(context.Background(), "container_id", "10")
	if err != nil {
		t.Fatal(err)
	}
}
