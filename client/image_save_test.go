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

func TestImageSaveServerError(t *testing.T) {
	expectedError := "Server error"

	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, expectedError)),
	}

	_, err := client.ImageSave(context.Background(), "test_image_save_500")
	if err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("expected (%v), got (%v)", expectedError, err)
	}
}

func TestImageSaveOK(t *testing.T) {
	expectedImageName := "test_image_save_ok"
	expectedURL := "/images/save"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}

		if req.Method != "GET" {
			return nil, fmt.Errorf("expected GET method, got %s", req.Method)
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

	if _, err := client.ImageSave(context.Background(), expectedImageName); err != nil {
		t.Fatal(err)
	}
}
