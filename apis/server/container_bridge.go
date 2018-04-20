package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *Server) removeContainers(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	option := &mgr.ContainerRemoveOption{
		Force: httputils.BoolValue(req, "force"),
		// TODO Volume and Link will be supported in the future.
		Volume: httputils.BoolValue(req, "v"),
		Link:   httputils.BoolValue(req, "link"),
	}

	if err := s.ContainerMgr.Remove(ctx, name, option); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) renameContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	oldName := mux.Vars(req)["name"]
	newName := req.FormValue("name")

	if err := s.ContainerMgr.Rename(ctx, oldName, newName); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) restartContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	var (
		t   int
		err error
	)

	if v := req.FormValue("t"); v != "" {
		if t, err = strconv.Atoi(v); err != nil {
			return httputils.NewHTTPError(err, http.StatusBadRequest)
		}
	}

	name := mux.Vars(req)["name"]

	if err = s.ContainerMgr.Restart(ctx, name, int64(t)); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

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

	var attach *mgr.AttachConfig

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

	return s.ContainerMgr.StartExec(ctx, name, config, attach)
}

func (s *Server) getExecInfo(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]
	execInfo, err := s.ContainerMgr.InspectExec(ctx, name)
	if err != nil {
		return err
	}
	return EncodeResponse(rw, http.StatusOK, execInfo)
}

func (s *Server) createContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.ContainerCreateConfig{}
	reader := req.Body
	var ex error
	if s.ContainerPlugin != nil {
		logrus.Infof("invoke container pre-create hook in plugin")
		if reader, ex = s.ContainerPlugin.PreCreate(req.Body); ex != nil {
			return errors.Wrapf(ex, "pre-create plugin point execute failed")
		}
	}
	// decode request body
	if err := json.NewDecoder(reader).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := req.FormValue("name")

	// to do compensation to potential nil pointer after validation
	if config.HostConfig == nil {
		config.HostConfig = &types.HostConfig{}
	}
	if config.NetworkingConfig == nil {
		config.NetworkingConfig = &types.NetworkingConfig{}
	}

	container, err := s.ContainerMgr.Create(ctx, name, config)
	if err != nil {
		return err
	}

	return EncodeResponse(rw, http.StatusCreated, container)
}

func (s *Server) startContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	detachKeys := req.FormValue("detachKeys")

	if err := s.ContainerMgr.Start(ctx, name, detachKeys); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) stopContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	var (
		t   int
		err error
	)

	if v := req.FormValue("t"); v != "" {
		if t, err = strconv.Atoi(v); err != nil {
			return httputils.NewHTTPError(err, http.StatusBadRequest)
		}
	}

	name := mux.Vars(req)["name"]

	if err = s.ContainerMgr.Stop(ctx, name, int64(t)); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) pauseContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	if err := s.ContainerMgr.Pause(ctx, name); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) unpauseContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	if err := s.ContainerMgr.Unpause(ctx, name); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) attachContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	_, upgrade := req.Header["Upgrade"]

	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		return fmt.Errorf("not a hijack connection, container: %s", name)
	}

	attach := &mgr.AttachConfig{
		Hijack:  hijacker,
		Stdin:   req.FormValue("stdin") == "1",
		Stdout:  true,
		Stderr:  true,
		Upgrade: upgrade,
	}

	if err := s.ContainerMgr.Attach(ctx, name, attach); err != nil {
		// TODO handle error
	}

	return nil
}

func (s *Server) getContainers(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	option := &mgr.ContainerListOption{
		All: httputils.BoolValue(req, "all"),
	}

	metas, err := s.ContainerMgr.List(ctx, func(meta *mgr.ContainerMeta) bool {
		return true
	}, option)
	if err != nil {
		return err
	}

	containerList := make([]types.Container, 0, len(metas))

	for _, m := range metas {
		status, err := m.FormatStatus()
		if err != nil {
			return err
		}

		t, err := time.Parse(utils.TimeLayout, m.Created)
		if err != nil {
			return err
		}

		container := types.Container{
			ID:         m.ID,
			Names:      []string{m.Name},
			Image:      m.Config.Image,
			Command:    strings.Join(m.Config.Cmd, " "),
			Status:     status,
			Created:    t.UnixNano(),
			Labels:     m.Config.Labels,
			HostConfig: m.HostConfig,
		}

		if m.NetworkSettings != nil {
			container.NetworkSettings = &types.ContainerNetworkSettings{
				Networks: m.NetworkSettings.Networks,
			}
		}

		containerList = append(containerList, container)
	}
	return EncodeResponse(rw, http.StatusOK, containerList)
}

func (s *Server) getContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	meta, err := s.ContainerMgr.Get(ctx, name)
	if err != nil {
		return err
	}

	container := types.ContainerJSON{
		ID:         meta.ID,
		Name:       meta.Name,
		Image:      meta.Config.Image,
		Created:    meta.Created,
		State:      meta.State,
		Config:     meta.Config,
		HostConfig: meta.HostConfig,
		GraphDriver: &types.GraphDriverData{
			Name: "overlay2",
			Data: map[string]string{
				"BaseFS": meta.BaseFS,
			},
		},
	}

	if meta.NetworkSettings != nil {
		container.NetworkSettings = &types.NetworkSettings{
			Networks: meta.NetworkSettings.Networks,
		}
	}

	container.Mounts = []types.MountPoint{}
	for _, mp := range meta.Mounts {
		container.Mounts = append(container.Mounts, *mp)
	}

	return EncodeResponse(rw, http.StatusOK, container)
}

func (s *Server) updateContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.UpdateConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := mux.Vars(req)["name"]

	if err := s.ContainerMgr.Update(ctx, name, config); err != nil {
		return httputils.NewHTTPError(err, http.StatusInternalServerError)
	}

	rw.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) upgradeContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.ContainerUpgradeConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := mux.Vars(req)["name"]

	if err := s.ContainerMgr.Upgrade(ctx, name, config); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) topContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	procList, err := s.ContainerMgr.Top(ctx, name, req.Form.Get("ps_args"))
	if err != nil {
		return err
	}

	return EncodeResponse(rw, http.StatusOK, procList)
}

func (s *Server) logsContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	//opts := &types.ContainerLogsOptions{}

	// TODO
	return nil
}

func (s *Server) resizeContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
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

	if err := s.ContainerMgr.Resize(ctx, name, opts); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusOK)
	return nil
}
