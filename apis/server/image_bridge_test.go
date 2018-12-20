package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/stretchr/testify/assert"
)

type mockImgePull struct {
	mgr.ImageMgr
	handler func(ctx context.Context, imageRef string, authConfig *types.AuthConfig, out io.Writer) error
}

func (m *mockImgePull) PullImage(ctx context.Context, imageRef string, authConfig *types.AuthConfig, out io.Writer) error {
	return m.handler(ctx, imageRef, authConfig, out)
}

func Test_pullImage_without_tag(t *testing.T) {
	var s Server

	s.ImageMgr = &mockImgePull{
		ImageMgr: &mgr.ImageManager{},
		handler: func(ctx context.Context, imageRef string, authConfig *types.AuthConfig, out io.Writer) error {
			assert.Equal(t, "reg.abc.com/base/os:7.2", imageRef)
			return nil
		},
	}
	req := &http.Request{
		Form:   map[string][]string{"fromImage": {"reg.abc.com/base/os:7.2"}},
		Header: map[string][]string{},
	}
	s.pullImage(context.Background(), nil, req)
}

func Test_pullImage_counter(t *testing.T) {
	var s Server
	ctx := context.Background()

	key := fmt.Sprintf(`engine_daemon_image_actions_counter_total{action="%s"}`, "pull")
	keySuccess := fmt.Sprintf(`engine_daemon_image_success_actions_counter_total{action="%s"}`, "pull")
	countBefore, countSuccessBefore := getMetric(ctx, t, &s, key, keySuccess)

	ch := make(chan int, 1)
	go func() {
		s.ImageMgr = &mockImgePull{
			ImageMgr: &mgr.ImageManager{},
			handler: func(ctx context.Context, imageRef string, authConfig *types.AuthConfig, out io.Writer) error {
				assert.Equal(t, "reg.abc.com/base/os:7.2", imageRef)
				time.Sleep(2 * time.Second)
				return nil
			},
		}
		req := &http.Request{
			Form:   map[string][]string{"fromImage": {"reg.abc.com/base/os:7.2"}},
			Header: map[string][]string{},
		}
		s.pullImage(ctx, nil, req)

		ch <- 1
	}()

	countMedium, countSuccessMedium := getMetric(ctx, t, &s, key, keySuccess)
	assert.Equal(t, countBefore, countMedium)
	assert.Equal(t, countSuccessBefore, countSuccessMedium)
	<-ch

	count, successCount := getMetric(ctx, t, &s, key, keySuccess)
	assert.Equal(t, countBefore+1, count)
	assert.Equal(t, countSuccessBefore+1, successCount)
}

func getMetric(ctx context.Context, t *testing.T, s *Server, key string, keySuccess string) (int, int) {
	req := httptest.NewRequest("get", "/metrics", nil)
	w := httptest.NewRecorder()
	err := s.metrics(ctx, w, req)
	assert.Equal(t, nil, err)
	resp := w.Result()
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	value := ""
	valueSuccess := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, key) {
			kv := strings.Split(line, " ")
			if len(kv) == 2 {
				value = kv[1]
			}
		} else if strings.Contains(line, keySuccess) {
			kv := strings.Split(line, " ")
			if len(kv) == 2 {
				valueSuccess = kv[1]
			}
		}
	}

	iCount := 0
	if value != "" {
		iCount, err = strconv.Atoi(value)
		assert.Equal(t, nil, err)
	}

	iCountSuccess := 0
	if valueSuccess != "" {
		iCountSuccess, err = strconv.Atoi(valueSuccess)
		assert.Equal(t, nil, err)
	}

	return iCount, iCountSuccess
}
