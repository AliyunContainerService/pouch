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

func TestNetworkDisconnectNotFoundError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusNotFound, "Not Found")),
	}
	err := client.NetworkDisconnect(context.Background(), "no network", "no container", false)
	if err == nil || !strings.Contains(err.Error(), "Not Found") {
		t.Fatalf("expected a Not Found Error, got %v", err)
	}
}

func TestNetworkDisconnect(t *testing.T) {
	expectedURL := "/networks/networkID/disconnect"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Method != "POST" {
			return nil, fmt.Errorf("expected POST method, got %s", req.Method)
		}

		disconnect := &types.NetworkDisconnect{}
		if err := json.NewDecoder(req.Body).Decode(&disconnect); err != nil {
			return nil, err
		}

		if disconnect.Container != "containerID" {
			return nil, fmt.Errorf("expected 'containerID', got %s", disconnect.Container)
		}

		if disconnect.Force {
			return nil, fmt.Errorf("expected Force to be false, got %v", disconnect.Force)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	err := client.NetworkDisconnect(context.Background(), "networkID", "containerID", false)
	if err != nil {
		t.Fatal(err)
	}
}
