package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/metrics"
	"github.com/alibaba/pouch/apis/types"
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
		if index := strings.LastIndex(image, ":"); index > 0 {
			tag = image[index+1:]
			image = image[:index]
		}
	}
	// record the time spent during image pull procedure.
	defer func(start time.Time) {
		metrics.ImagePullSummary.WithLabelValues(image + ":" + tag).Observe(metrics.SinceInMicroseconds(start))
	}(time.Now())

	// get registry auth from Request header
	authStr := req.Header.Get("X-Registry-Auth")
	authConfig := types.AuthConfig{}
	if authStr != "" {
		data := base64.NewDecoder(base64.URLEncoding, strings.NewReader(authStr))
		if err := json.NewDecoder(data).Decode(&authConfig); err != nil {
			return err
		}
	}
	// Error information has be sent to client, so no need call resp.Write
	if err := s.ImageMgr.PullImage(ctx, image+":"+tag, &authConfig, rw); err != nil {
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
		return meta.Image == image.ID
	}, &mgr.ContainerListOption{All: true})
	if err != nil {
		return err
	}

	isForce := httputils.BoolValue(req, "force")
	if !isForce && len(containers) > 0 {
		return fmt.Errorf("Unable to remove the image %q (must force) - container %s is using this image", image.ID, containers[0].ID)
	}

	option := &mgr.ImageRemoveOption{
		Force: isForce,
	}

	if err := s.ImageMgr.RemoveImage(ctx, image, name, option); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}
