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

func TestDaemonUpdateError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	err := client.DaemonUpdate(context.Background(), &types.DaemonUpdateConfig{})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestDaemonUpdate(t *testing.T) {
	expectedURL := "/daemon/update"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
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

	if err := client.DaemonUpdate(context.Background(), &types.DaemonUpdateConfig{Labels: []string{"abc=def"}}); err != nil {
		t.Fatal(err)
	}
}
