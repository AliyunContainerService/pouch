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

func TestImageTagNotFoundError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusNotFound, "Not Found")),
	}

	err := client.ImageTag(context.Background(), "oops", "whatever")
	if err == nil || !strings.Contains(err.Error(), "Not Found") {
		t.Fatalf("expected a not found error, got %v", err)
	}
}

func TestImageTagOK(t *testing.T) {
	expectedURL := "/images/imagetagok/tag"

	expectedRepo, expectedTag := "pouch", "0.5.0"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}

		if req.Method != http.MethodPost {
			return nil, fmt.Errorf("expected POST method, got %s", req.Method)
		}

		if got := req.FormValue("repo"); got != expectedRepo {
			return nil, fmt.Errorf("expected repo is %s, got %s", expectedRepo, got)
		}

		if got := req.FormValue("tag"); got != expectedTag {
			return nil, fmt.Errorf("expected tag is %s, got %s", expectedTag, got)
		}

		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	err := client.ImageTag(context.Background(), "imagetagok", fmt.Sprintf("%s:%s", expectedRepo, expectedTag))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
