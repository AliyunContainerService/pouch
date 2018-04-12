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

func TestContainerResizeError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	err := client.ContainerResize(context.Background(), "nothing", "", "")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerResize(t *testing.T) {
	expectedURL := "/containers/container_id/resize"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		height := req.URL.Query().Get("h")
		if height != "200" {
			return nil, fmt.Errorf("height not set in URL query properly. Expected '200', got %s", height)
		}
		width := req.URL.Query().Get("w")
		if width != "300" {
			return nil, fmt.Errorf("width not set in URL query properly. Expected '300', got %s", width)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	err := client.ContainerResize(context.Background(), "container_id", "200", "300")
	if err != nil {
		t.Fatal(err)
	}
}
