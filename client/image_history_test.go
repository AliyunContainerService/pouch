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

func TestImageHistoryServerError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ImageHistory(context.Background(), "test_image_history_500")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestImageHistoryNotFoundError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusNotFound, "Not Found")),
	}
	_, err := client.ImageHistory(context.Background(), "no image")
	if err == nil || !strings.Contains(err.Error(), "Not Found") {
		t.Fatalf("expected a Not Found Error, got %v", err)
	}
}

func TestImageHistory(t *testing.T) {
	expectedURL := "/images/image_id/history"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", req.Method)
		}

		imageHistoryResp, err := json.Marshal([]types.HistoryResultItem{
			{
				ID:   "1",
				Size: int64(94),
			},
			{
				ID:   "2",
				Size: int64(703),
			},
		})
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(imageHistoryResp))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	history, err := client.ImageHistory(context.Background(), "image_id")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, history[0].ID, "1")
	assert.Equal(t, history[0].Size, int64(94))

	assert.Equal(t, history[1].ID, "2")
	assert.Equal(t, history[1].Size, int64(703))
}
