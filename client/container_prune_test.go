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
	"github.com/alibaba/pouch/pkg/utils/filters"

	"github.com/stretchr/testify/assert"
)

func TestContainerPruneError(t *testing.T) {
	client := &APIClient{
		HTTPCli: newMockClient(errorMockResponse(http.StatusInternalServerError, "Server error")),
	}

	var filter = make(map[string][]string)
	filter["id"] = []string{"sdf", "saf", "adsf"}

	_, err := client.ContainerPrune(context.Background(), nil)

	if err == nil || err.Error() == "Server error" {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerPruneList(t *testing.T) {
	expectedURL := "/containers/prune"

	httpClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, expectedURL) {
			return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
		}
		_, err := filters.FromURLParam(req.FormValue("filters"))
		if err != nil {
			return nil, err
		}
		containersPruneJSON := types.ContainerPruneResp{
			ContainersDeleted: []string{
				"4cedc676735f7802c270d6025afce9e52f233c0216dd40c52d676e8756a0038d",
				"120305964b9c7a9341faeb49957fc8de772ee59e1ae2a22f1bd476abfd364062",
				"b304e9d23cae2004915ce0f55cb662e1bce1d1dfa04ba8c36899e00dc65f6f7e",
				"c6207bd7914bd7bb9c8af00ba87060eec747cae7485dc8551ec1c9e453903da9",
			},
			SpaceReclaimed: 34,
		}
		b, err := json.Marshal(containersPruneJSON)
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
	containerPruneResp, err := client.ContainerPrune(context.Background(), make(map[string][]string))
	if err != nil {
		t.Fatal(err)
	}
	if len(containerPruneResp.ContainersDeleted) != 4 {
		t.Fatalf("expected 4 containers, got %v", containerPruneResp.ContainersDeleted)
	}
	assert.Equal(t, containerPruneResp.SpaceReclaimed, int64(34))
}
