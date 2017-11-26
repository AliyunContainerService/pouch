package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alibaba/pouch/apis/metrics"

	"github.com/sirupsen/logrus"
)

// pullImage will pull an image from a specified registry.
func (s *Server) pullImage(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	image := req.FormValue("fromImage")
	tag := req.FormValue("tag")

	if image == "" {
		resp.WriteHeader(http.StatusBadRequest)
		return fmt.Errorf("fromImage cannot be empty")
	}

	if tag == "" {
		tag = "latest"
	}
	// record the time spent during image pull procedure.
	defer func(start time.Time) {
		metrics.ImagePullSummary.WithLabelValues(image + ":" + tag).Observe(metrics.SinceInMicroseconds(start))
	}(time.Now())

	// Error information has be sent to client, so no need call resp.Write
	if err := s.ImageMgr.PullImage(ctx, image, tag, resp); err != nil {
		logrus.Errorf("failed to pull image %s:%s: %v", image, tag, err)
		return nil
	}

	return nil
}

func (s *Server) listImages(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	filters := req.FormValue("failters")

	imageList, err := s.ImageMgr.ListImages(ctx, filters)
	if err != nil {
		logrus.Errorf("failed to list images in containerd: %v", err)
		resp.WriteHeader(http.StatusBadRequest)
		return err
	}
	return json.NewEncoder(resp).Encode(imageList)
}

func (s *Server) searchImages(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	searchPattern := req.FormValue("term")
	registry := req.FormValue("registry")

	response, err := s.ImageMgr.SearchImages(ctx, searchPattern, registry)
	if err != nil {
		logrus.Errorf("failed to search images from resgitry: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return json.NewEncoder(resp).Encode(response)
}
