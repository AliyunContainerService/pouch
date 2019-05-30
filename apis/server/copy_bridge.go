package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/alibaba/pouch/pkg/httputils"

	"github.com/gorilla/mux"
)

func (s *Server) putContainersArchive(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]
	path := req.FormValue("path")

	if name == "" {
		return httputils.NewHTTPError(errors.New("name can't be empty"), http.StatusBadRequest)
	}
	if path == "" {
		return httputils.NewHTTPError(errors.New("path can't be empty"), http.StatusBadRequest)
	}
	noOverwriteDirNonDir := httputils.BoolValue(req, "noOverwriteDirNonDir")
	copyUIDGID := httputils.BoolValue(req, "copyUIDGID")

	return s.ContainerMgr.ExtractToDir(ctx, name, path, copyUIDGID, noOverwriteDirNonDir, req.Body)
}

func (s *Server) headContainersArchive(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]
	path := req.FormValue("path")

	if name == "" {
		return httputils.NewHTTPError(errors.New("name can't be empty"), http.StatusBadRequest)
	}
	if path == "" {
		return httputils.NewHTTPError(errors.New("path can't be empty"), http.StatusBadRequest)
	}

	stat, err := s.ContainerMgr.StatPath(ctx, name, path)
	if err != nil {
		return err
	}

	statJSON, err := json.Marshal(stat)
	if err != nil {
		return err
	}

	rw.Header().Set(
		"X-Docker-Container-Path-Stat",
		base64.StdEncoding.EncodeToString(statJSON),
	)

	return nil
}

func (s *Server) getContainersArchive(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]
	path := req.FormValue("path")

	if name == "" {
		return httputils.NewHTTPError(errors.New("name can't be empty"), http.StatusBadRequest)
	}
	if path == "" {
		return httputils.NewHTTPError(errors.New("path can't be empty"), http.StatusBadRequest)
	}

	tarArchive, stat, err := s.ContainerMgr.ArchivePath(ctx, name, path)
	if err != nil {
		return err
	}
	defer tarArchive.Close()

	statJSON, err := json.Marshal(stat)
	if err != nil {
		return err
	}
	rw.Header().Set(
		"X-Docker-Container-Path-Stat",
		base64.StdEncoding.EncodeToString(statJSON),
	)

	rw.Header().Set("Content-Type", "application/x-tar")
	_, err = io.Copy(rw, tarArchive)

	return err
}
