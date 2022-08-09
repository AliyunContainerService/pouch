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

func TestContainerTopError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ContainerTop(context.Background(), "nothing", []string{})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerTop(t *testing.T) {
	expectedURL := "/containers/container_id/top"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		query := req.URL.Query()
		args := query.Get("ps_args")
		if args != "arg1 arg2" {
			return nil, fmt.Errorf("args not set in URL properly. Expected 'arg1 arg2', got %s", args)
		}
		psListJSON := types.ContainerProcessList{
			Processes: [][]string{{"bar11", "bar12"}, {"bar21", "bar22"}},
			Titles:    []string{"foo1", "foo2"},
		}
		b, err := json.Marshal(psListJSON)
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
	psList, err := client.ContainerTop(context.Background(), "container_id", []string{"arg1", "arg2"})
	if err != nil {
		t.Fatal(err)
	}

	if len(psList.Titles) != 2 {
		t.Fatalf("expected 2 titles, got %v", len(psList.Titles))
	}
	for _, ps := range psList.Processes {
		if len(ps) != len(psList.Titles) {
			t.Fatalf("expected 2 values, got %v", len(ps))
		}
	}
}
