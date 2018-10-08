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

func TestCommitError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ContainerCommit(context.Background(), "nothing", types.ContainerCommitOptions{})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestCommit(t *testing.T) {
	expectedURL := "/commit"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Header.Get("Content-Type") != "application/json" {
			return nil, fmt.Errorf("expected application/json set in header")
		}
		options := types.ContainerCommitOptions{}
		if err := json.NewDecoder(req.Body).Decode(&options); err != nil {
			return nil, fmt.Errorf("failed to parse json: %v", err)
		}

		if options.Repository != "foo" {
			return nil, fmt.Errorf("expected Repository %s, obtain %s", "foo", options.Repository)
		}
		if options.Tag != "bar" {
			return nil, fmt.Errorf("expected Tag %s, obtain %s", "bar", options.Tag)
		}

		resp := types.ContainerCommitResp{
			ID: "newid",
		}
		b, err := json.Marshal(resp)
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

	r, err := client.ContainerCommit(context.Background(), "id", types.ContainerCommitOptions{
		Repository: "foo",
		Tag:        "bar"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, r.ID, "newid")
}
