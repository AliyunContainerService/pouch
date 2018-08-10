package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/alibaba/pouch/apis/types"
)

func TestCheckpointCreateError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	err := client.ContainerCheckpointCreate(context.Background(), "nothing", types.CheckpointCreateOptions{CheckpointID: "cp0"})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestCheckpointCreate(t *testing.T) {
	expectedURL := "/containers/id/checkpoints"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Header.Get("Content-Type") == "application/json" {
			options := types.CheckpointCreateOptions{}
			if err := json.NewDecoder(req.Body).Decode(&options); err != nil {
				return nil, fmt.Errorf("failed to parse json: %v", err)
			}

			if options.CheckpointID != "cp0" {
				return nil, fmt.Errorf("expected CheckpointID %s, obtain %s", "cp0", options.CheckpointID)
			}
		}

		return &http.Response{
			StatusCode: http.StatusOK,
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	err := client.ContainerCheckpointCreate(context.Background(), "id", types.CheckpointCreateOptions{CheckpointID: "cp0"})
	if err != nil {
		t.Fatal(err)
	}
}
