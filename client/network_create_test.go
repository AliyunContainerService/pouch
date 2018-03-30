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

func TestNetworkCreateServerError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.NetworkCreate(context.Background(), &types.NetworkCreateConfig{})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestNetworkCreateDuplicatedError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusConflict, "Container already exists")),
	}
	_, err := client.NetworkCreate(context.Background(), &types.NetworkCreateConfig{})
	if err == nil || strings.Contains(err.Error(), "duplicated container") {
		t.Fatalf("expected a Container Already Exists Error, got %v", err)
	}
}

func TestNetworkCreate(t *testing.T) {
	expectedURL := "/networks/create"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Header.Get("Content-Type") == "application/json" {
			createConfig := types.NetworkCreateConfig{}
			if err := json.NewDecoder(req.Body).Decode(&createConfig); err != nil {
				return nil, fmt.Errorf("failed to parse json: %v", err)
			}
		}

		netCreateResp, err := json.Marshal(types.NetworkCreateResp{
			ID:      "network_id",
			Warning: "warning",
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(netCreateResp))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	res, err := client.NetworkCreate(context.Background(), &types.NetworkCreateConfig{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, res.ID, "network_id")
	assert.Equal(t, res.Warning, "warning")
}
