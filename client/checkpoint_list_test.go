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

func TestCheckpointListError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ContainerCheckpointList(context.Background(), "nothing", types.CheckpointListOptions{CheckpointDir: "nothing"})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestCheckpointList(t *testing.T) {
	expectedURL := "/containers/container_id/checkpoints"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		list := make([]string, 0)
		b, err := json.Marshal(&list)
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
	if _, err := client.ContainerCheckpointList(context.Background(), "container_id", types.CheckpointListOptions{CheckpointDir: "nothing"}); err != nil {
		t.Fatal(err)
	}
}
