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

func TestContainerGetError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ContainerGet(context.Background(), "nothing")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerGet(t *testing.T) {
	expectedURL := "/containers/container_id/json"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		containerJSON := types.ContainerJSON{
			Driver: "Driver",
			Image:  "Image",
		}
		b, err := json.Marshal(containerJSON)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(b))),
		}, nil
	})
	client := &APIClient{
		HTTPCli: httpClient,
	}
	if _, err := client.ContainerGet(context.Background(), "container_id"); err != nil {
		t.Fatal(err)
	}
}
