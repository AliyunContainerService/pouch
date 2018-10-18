package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/httputils"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s *Server) createContainerExec(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.ExecCreateConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := mux.Vars(req)["name"]

	logCreateOptions("container exec for "+name, config)

	id, err := s.ContainerMgr.CreateExec(ctx, name, config)
	if err != nil {
		return err
	}

	execCreateResp := &types.ExecCreateResp{
		ID: id,
	}

	return EncodeResponse(rw, http.StatusCreated, execCreateResp)
}

func (s *Server) startContainerExec(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.ExecStartConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := mux.Vars(req)["name"]
	_, upgrade := req.Header["Upgrade"]

	ba, _ := json.Marshal(config)
	logrus.Infof("start exec %s, upgrade: %v, body: %s", name, upgrade, string(ba))

	if err := s.ContainerMgr.CheckExecExist(ctx, name); err != nil {
		return err
	}

	var attach *mgr.AttachConfig

	// TODO(huamin.thm): support detach exec process through http post method
	if !config.Detach {
		hijacker, ok := rw.(http.Hijacker)
		if !ok {
			return fmt.Errorf("not a hijack connection, container: %s", name)
		}

		attach = &mgr.AttachConfig{
			Hijack:  hijacker,
			Stdin:   config.Tty,
			Stdout:  true,
			Stderr:  true,
			Upgrade: upgrade,
		}
	}

	if err := s.ContainerMgr.StartExec(ctx, name, attach); err != nil {
		logrus.Errorf("failed to run exec process: %s", err)
	}

	return nil
}

func (s *Server) getExecInfo(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]
	execInfo, err := s.ContainerMgr.InspectExec(ctx, name)
	if err != nil {
		return err
	}
	return EncodeResponse(rw, http.StatusOK, execInfo)
}

func (s *Server) resizeExec(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	height, err := strconv.Atoi(req.FormValue("h"))
	if err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	width, err := strconv.Atoi(req.FormValue("w"))
	if err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	opts := types.ResizeOptions{
		Height: int64(height),
		Width:  int64(width),
	}

	name := mux.Vars(req)["name"]

	if err := s.ContainerMgr.ResizeExec(ctx, name, opts); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusOK)
	return nil

}
