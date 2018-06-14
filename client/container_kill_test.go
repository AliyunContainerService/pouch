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

func TestContainerKillError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	err := client.ContainerKill(context.Background(), "nothing", "SIGKILL")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerKillNotFoundError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusNotFound, "Not Found")),
	}
	err := client.ContainerKill(context.Background(), "no container", "SIGKILL")
	if err == nil || !strings.Contains(err.Error(), "Not Found") {
		t.Fatalf("expected a Not Found Error, got %v", err)
	}
}

func TestContainerKill(t *testing.T) {
	expectedURL := "/containers/container_id/kill"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		signal := req.URL.Query().Get("signal")
		if signal != "SIGKILL" {
			return nil, fmt.Errorf("signal not set in URL query properly. Expected 'SIGKILL', got %s", signal)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	err := client.ContainerKill(context.Background(), "container_id", "SIGKILL")
	if err != nil {
		t.Fatal(err)
	}
}
