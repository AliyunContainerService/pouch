package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/alibaba/pouch/apis/types"
)

func TestContainerLogsServerError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}

	_, err := client.ContainerLogs(context.Background(), "nothing", types.ContainerLogsOptions{})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerLogsOK(t *testing.T) {
	expectedURL := "/containers/container_id/logs"
	expectedSinceTS := "1531728000.000000000" // 2018-07-16T08:00Z
	expectedUntilTS := "1531728300.000000000" // 2018-07-16T08:05Z

	opts := types.ContainerLogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: false,
		Timestamps: true,

		Since: "2018-07-16T08:00Z",
		Until: "2018-07-16T08:05Z",
		Tail:  "10",
	}

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL %s, got %s", expectedURL, req.URL)
		}

		if req.Method != http.MethodGet {
			return nil, fmt.Errorf("expected HTTP Method = %s, got %s", http.MethodGet, req.Method)
		}

		query := req.URL.Query()
		if got := query.Get("follow"); got != "1" {
			return nil, fmt.Errorf("expected follow mode (1), got %v", got)
		}

		if got := query.Get("stdout"); got != "1" {
			return nil, fmt.Errorf("expected stdout mode (1), got %v", got)
		}

		if got := query.Get("stderr"); got != "" {
			return nil, fmt.Errorf("expected without stderr mode, got %v", got)
		}

		if got := query.Get("timestamps"); got != "1" {
			return nil, fmt.Errorf("expected timestamps mode, got %v", got)
		}

		if got := query.Get("tail"); got != "10" {
			return nil, fmt.Errorf("expected tail = %v, got %v", opts.Tail, got)
		}

		if got := query.Get("since"); got != expectedSinceTS {
			return nil, fmt.Errorf("expected since = %v, got %v", expectedSinceTS, got)
		}

		if got := query.Get("until"); got != expectedUntilTS {
			return nil, fmt.Errorf("expected since = %v, got %v", expectedUntilTS, got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil
	})

	client := &APIClient{
		HTTPCli: httpClient,
	}

	_, err := client.ContainerLogs(context.Background(), "container_id", opts)
	if err != nil {
		t.Fatal(err)
	}
}
