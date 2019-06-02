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

	"github.com/stretchr/testify/assert"
)

func TestNetworkPruneError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}

	_, err := client.NetworkPrune(context.Background())

	if err == nil || err.Error() == "Server error" {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestNetworkPruneList(t *testing.T) {
	expectedURL := "/networks/prune"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}

		networksPruneJSON := []string{
			"network0",
			"network1",
			"network2",
		}

		b, err := json.Marshal(networksPruneJSON)
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
	networkPruneResp, err := client.NetworkPrune(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(networkPruneResp) != 3 {
		t.Fatalf("expected 3 networks, got %v", networkPruneResp)
	}
	assert.Equal(t, networkPruneResp[0], "network0")
	assert.Equal(t, networkPruneResp[1], "network1")
	assert.Equal(t, networkPruneResp[2], "network2")
}
