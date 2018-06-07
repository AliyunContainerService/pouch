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

func TestContainerCreateExecError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ContainerCreateExec(context.Background(), "nothing", &types.ExecCreateConfig{})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerCreateExec(t *testing.T) {
	expectedURL := "/containers/container_id/exec"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Header.Get("Content-Type") == "application/json" {
			createExecConfig := &types.ExecCreateConfig{}
			if err := json.NewDecoder(req.Body).Decode(createExecConfig); err != nil {
				return nil, fmt.Errorf("failed to parse json: %v", err)
			}
		}
		createExecResponse := types.ExecCreateResp{
			ID: "container_id",
		}
		b, err := json.Marshal(createExecResponse)
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

	res, err := client.ContainerCreateExec(context.Background(), "container_id", &types.ExecCreateConfig{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, res.ID, "container_id")
}

func TestContainerInspectExecError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ContainerExecInspect(context.Background(), "nothing")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerInspectExec(t *testing.T) {
	expectedURL := "/exec/container_id/json"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", req.Method)
		}
		b, err := json.Marshal(types.ContainerExecInspect{
			ID:          "exec_id",
			ContainerID: "container_id",
		})
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

	res, err := client.ContainerExecInspect(context.Background(), "container_id")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, res.ID, "exec_id")
	assert.Equal(t, res.ContainerID, "container_id")
}
