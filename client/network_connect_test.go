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

func TestNetworkConnectNotFoundError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusNotFound, "Not Found")),
	}
	err := client.NetworkConnect(context.Background(), "no_network", &types.NetworkConnect{})
	if err == nil || !strings.Contains(err.Error(), "Not Found") {
		t.Fatalf("expected a Not Found Error, got %v", err)
	}
}

func TestNetworkConnect(t *testing.T) {
	expectedURL := "/networks/network_id/connect"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Method != "POST" {
			return nil, fmt.Errorf("expected POST method, got %s", req.Method)
		}

		connect := &types.NetworkConnect{}
		if err := json.NewDecoder(req.Body).Decode(connect); err != nil {
			return nil, err
		}

		if connect.Container != "container_id" {
			return nil, fmt.Errorf("expected 'containerID', got %s", connect.Container)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	err := client.NetworkConnect(context.Background(), "network_id", &types.NetworkConnect{Container: "container_id"})
	if err != nil {
		t.Fatal(err)
	}
}
