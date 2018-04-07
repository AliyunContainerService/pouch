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

func TestContainerRemoveError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	err := client.ContainerRemove(context.Background(), "nothing", true)
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerRemoveNotFoundError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusNotFound, "Not Found")),
	}
	err := client.ContainerRemove(context.Background(), "no contaienr", true)
	if err == nil || !strings.Contains(err.Error(), "Not Found") {
		t.Fatalf("expected a Not Found Error, got %v", err)
	}
}

func TestContainerRemove(t *testing.T) {
	expectedURL := "/containers/container_id"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		force := req.URL.Query().Get("force")
		if force != "true" {
			return nil, fmt.Errorf("force not set in URL properly. Expected 'true', got %s", force)
		}
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})
	client := &APIClient{
		HTTPCli: httpClient,
	}
	err := client.ContainerRemove(context.Background(), "container_id", true)
	if err != nil {
		t.Fatal(err)
	}
}
