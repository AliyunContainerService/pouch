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

func TestContainerCreateError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ContainerCreate(context.Background(), types.ContainerConfig{}, nil, nil, "nothing")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerCreate(t *testing.T) {
	expectedURL := "/containers/create"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Header.Get("Content-Type") == "application/json" {
			createConfig := types.ContainerCreateConfig{}
			if err := json.NewDecoder(req.Body).Decode(&createConfig); err != nil {
				return nil, fmt.Errorf("failed to parse json: %v", err)
			}
		}

		name := req.URL.Query().Get("name")
		if name != "container_name" {
			return nil, fmt.Errorf("container name not set in URL query properly. Expected `container_name`, got %s", name)
		}
		containerCreateResp := types.ContainerCreateResp{
			ID:   "container_id",
			Name: "container_name",
		}
		b, err := json.Marshal(containerCreateResp)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(b))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	container, err := client.ContainerCreate(context.Background(), types.ContainerConfig{}, &types.HostConfig{}, &types.NetworkingConfig{}, "container_name")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, container.ID, "container_id")
	assert.Equal(t, container.Name, "container_name")
}
