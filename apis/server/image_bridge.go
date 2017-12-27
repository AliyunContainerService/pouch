package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alibaba/pouch/apis/metrics"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/httputils"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// pullImage will pull an image from a specified registry.
func (s *Server) pullImage(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	image := req.FormValue("fromImage")
	tag := req.FormValue("tag")

	if image == "" {
		err := fmt.Errorf("fromImage cannot be empty")
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	if tag == "" {
		tag = "latest"
	}
	// record the time spent during image pull procedure.
	defer func(start time.Time) {
		metrics.ImagePullSummary.WithLabelValues(image + ":" + tag).Observe(metrics.SinceInMicroseconds(start))
	}(time.Now())

	// Error information has be sent to client, so no need call resp.Write
	if err := s.ImageMgr.PullImage(ctx, image, tag, rw); err != nil {
		logrus.Errorf("failed to pull image %s:%s: %v", image, tag, err)
		return nil
	}

	return nil
}

func (s *Server) listImages(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	filters := req.FormValue("filters")

	imageList, err := s.ImageMgr.ListImages(ctx, filters)
	if err != nil {
		logrus.Errorf("failed to list images: %v", err)
		return err
	}
	return EncodeResponse(rw, http.StatusOK, imageList)
}

func (s *Server) getImage(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	idOrRef := mux.Vars(req)["name"]

	imageInfo, err := s.ImageMgr.GetImage(ctx, idOrRef)
	if err != nil {
		logrus.Errorf("failed to get image: %v", err)
		return err
	}

	return EncodeResponse(rw, http.StatusOK, imageInfo)
}

func (s *Server) searchImages(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	searchPattern := req.FormValue("term")
	registry := req.FormValue("registry")

	searchResultItem, err := s.ImageMgr.SearchImages(ctx, searchPattern, registry)
	if err != nil {
		logrus.Errorf("failed to search images from resgitry: %v", err)
		return err
	}
	return EncodeResponse(rw, http.StatusOK, searchResultItem)
}

// removeImage deletes an image by reference.
func (s *Server) removeImage(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	image, err := s.ImageMgr.GetImage(ctx, name)
	if err != nil {
		return err
	}

	containers, err := s.ContainerMgr.List(ctx, func(meta *mgr.ContainerMeta) bool {
		return meta.Image == image.Name
	})
	if err != nil {
		return err
	}

	isForce := httputils.BoolValue(req, "force")
	if !isForce && len(containers) > 0 {
		return fmt.Errorf("Unable to remove the image %q (must force) - container %s is using this image", image.Name, containers[0].ID)
	}

	option := &mgr.ImageRemoveOption{
		Force: isForce,
	}

	if err := s.ImageMgr.RemoveImage(ctx, image, option); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}
