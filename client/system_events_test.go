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

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"

	"github.com/stretchr/testify/assert"
)

func TestEventsError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.Events(context.Background(), "", "", filters.Args{})
	if err == nil || !strings.Contains(err.Error(), "Server error") {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestEvents(t *testing.T) {
	expectedURL := "/events"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		events := types.EventsMessage{
			Action: "create",
			ID:     "abcd",
			Type:   types.EventTypeContainer,
			Actor: &types.EventsActor{
				ID:         "abcd",
				Attributes: map[string]string{"image": "busybox"},
			},
		}
		b, err := json.Marshal(events)
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

	body, err := client.Events(context.Background(), "", "", filters.Args{})
	if err != nil {
		t.Fatal(err)
	}

	dec := json.NewDecoder(body)
	var event types.EventsMessage
	if err := dec.Decode(&event); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, event.Action, "create")
	assert.Equal(t, event.ID, "abcd")
	assert.Equal(t, event.Type, types.EventTypeContainer)
}
