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

func TestRegistryLoginError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	loginConfig := types.AuthConfig{}
	_, err := client.RegistryLogin(context.Background(), &loginConfig)
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestRegistryLogin(t *testing.T) {
	expectedURL := "/auth"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("Expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		if req.Header.Get("Content-Type") == "application/json" {
			loginConfig := types.AuthConfig{}
			if err := json.NewDecoder(req.Body).Decode(&loginConfig); err != nil {
				return nil, fmt.Errorf("failed to parse json: %v", err)
			}
		}
		auth, err := json.Marshal(types.AuthResponse{
			IdentityToken: "aaa",
			Status:        "bbb",
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(auth))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	res, err := client.RegistryLogin(context.Background(), &types.AuthConfig{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, res.IdentityToken, "aaa")
	assert.Equal(t, res.Status, "bbb")
}
