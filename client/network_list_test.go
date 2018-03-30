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

func TestNetworkListServerError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.NetworkList(context.Background())
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestNetworkList(t *testing.T) {
	expectedURL := "/networks"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", req.Method)
		}

		netListResp, err := json.Marshal(types.NetworkListResp{
			Networks: []*types.NetworkInfo{
				{
					Driver: "bridge",
					ID:     "1",
					Name:   "net-1",
				},
				{
					Driver: "bridge",
					ID:     "2",
					Name:   "net-2",
				},
			},
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(netListResp))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	network, err := client.NetworkList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(network.Networks), 2)
}
