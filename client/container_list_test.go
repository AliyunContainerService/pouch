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

func TestContainerListError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ContainerList(context.Background(), true)
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerList(t *testing.T) {
	expectedURL := "/containers/json"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		query := req.URL.Query()
		all := query.Get("all")
		if all != "true" {
			return nil, fmt.Errorf("all not set in URL query properly. Expected '1', got %s", all)
		}
		containersJSON := []types.ContainerJSON{
			{
				Name:  "container1",
				Image: "Image1",
			},
			{
				Name:  "container1",
				Image: "Image1",
			},
		}
		b, err := json.Marshal(containersJSON)
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
	containers, err := client.ContainerList(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(containers) != 2 {
		t.Fatalf("expected 2 containers, got %v", containers)
	}
}
