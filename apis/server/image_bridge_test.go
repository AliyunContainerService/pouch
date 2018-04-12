package server

import (
	"context"
	"io"
	"net/http"
	"testing"

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
