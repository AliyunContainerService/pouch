package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestImagePullServerError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ImagePull(context.Background(), "image_name", "image_tag", "auth")
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestImagePullWrongError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusNotFound, "Image not found")),
	}
	_, err := client.ImagePull(context.Background(), "image_name", "image_tag", "auth")
	if err == nil || !strings.Contains(err.Error(), "Image not found") {
		t.Fatalf("expected an Image Not Found Error, got %v", err)
	}
}

func TestImagePull(t *testing.T) {
	expectedURL := "/images/create"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}

		if req.Method != "POST" {
			return nil, fmt.Errorf("expected POST method, got %s", req.Method)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	_, err := client.ImagePull(context.Background(), "image_name", "image_tag", "auth")
	if err != nil {
		t.Fatal(err)
	}

}
