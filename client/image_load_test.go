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

func TestImageLoadServerError(t *testing.T) {
	expectedError := "Server error"

	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, expectedError)),
	}

	err := client.ImageLoad(context.Background(), "test_image_load_500", nil)
	if err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("expected (%v), got (%v)", expectedError, err)
	}
}

func TestImageLoadOK(t *testing.T) {
	expectedURL := "/images/load"
	expectedImageName := "test_image_load_ok"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}

		if req.Method != "POST" {
			return nil, fmt.Errorf("expected POST method, got %s", req.Method)
		}

		if got := req.FormValue("name"); got != expectedImageName {
			return nil, fmt.Errorf("expected (%s), got %s", expectedImageName, got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	if err := client.ImageLoad(context.Background(), expectedImageName, nil); err != nil {
		t.Fatal(err)
	}

}
