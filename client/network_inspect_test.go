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

	"github.com/stretchr/testify/assert"
)

func TestNetworkInspectServerError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.NetworkInspect(context.Background(), "network_id")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestNetworkInspectNoFoundError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusConflict, "Not Found")),
	}
	_, err := client.NetworkInspect(context.Background(), "no network")
	if err == nil || !strings.Contains(err.Error(), "Not Found") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestNetworkInspect(t *testing.T) {
	expectedURL := "/networks/network_id"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", req.Method)
		}

		netInspectResp, err := json.Marshal(types.NetworkInspectResp{
			Driver:     "bridge",
			ID:         "1",
			Name:       "net-1",
			EnableIPV6: true,
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(netInspectResp))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	network, err := client.NetworkInspect(context.Background(), "network_id")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, network.Name, "net-1")
	assert.Equal(t, network.ID, "1")
	assert.Equal(t, network.EnableIPV6, true)
	assert.Equal(t, network.Driver, "bridge")
}
