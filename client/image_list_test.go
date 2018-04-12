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

func TestImageListServerError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ImageList(context.Background())
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestImageList(t *testing.T) {
	expectedURL := "/images"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", req.Method)
		}

		imageListResp, err := json.Marshal([]types.ImageInfo{
			{
				ID:   "1",
				Size: 703,
				Os:   "CentOS",
			},
			{
				ID:   "2",
				Size: 44,
				Os:   "Ubuntu TLS 16.04",
			},
		})
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(imageListResp))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	image, err := client.ImageList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(image), 2)
}
