package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/alibaba/pouch/apis/types"
)

func TestCheckpointDelError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	err := client.ContainerCheckpointDelete(context.Background(), "nothing", types.CheckpointDeleteOptions{CheckpointID: "noid"})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestCheckpointDel(t *testing.T) {
	expectedURL := "/containers/container_id/checkpoints/cp0"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		return &http.Response{
			StatusCode: http.StatusNoContent,
		}, nil
	})
	client := &APIClient{
		HTTPCli: httpClient,
	}
	err := client.ContainerCheckpointDelete(context.Background(), "container_id", types.CheckpointDeleteOptions{CheckpointID: "cp0"})
	if err != nil {
		t.Fatal(err)
	}
}
