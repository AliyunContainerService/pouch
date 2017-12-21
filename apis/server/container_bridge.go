package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/httputils"

	"github.com/gorilla/mux"
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
	oldName := mux.Vars(req)["id"]
	newName := req.FormValue("name")

	if err := s.ContainerMgr.Rename(ctx, oldName, newName); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) createContainerExec(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	config := &types.ExecCreateConfig{}

	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return err
	}

	if len(config.Cmd) == 0 {
		return fmt.Errorf("no exec process specified")
	}

	id, err := s.ContainerMgr.CreateExec(ctx, name, config)
	if err != nil {
		return err
	}

	execCreateResp := &types.ExecCreateResponse{
		ID: id,
	}

	return EncodeResponse(rw, http.StatusCreated, execCreateResp)
}

func (s *Server) startContainerExec(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	config := &types.ExecStartConfig{}

	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return err
	}

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
			Upgrade: true,
		}
	}

	return s.ContainerMgr.StartExec(ctx, name, config, attach)
}

func (s *Server) createContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	var config types.ContainerConfigWrapper

	if err := json.NewDecoder(req.Body).Decode(&config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := req.FormValue("name")

	container, err := s.ContainerMgr.Create(ctx, name, &config)
	if err != nil {
		return err
	}

	return EncodeResponse(rw, http.StatusCreated, container)
}

func (s *Server) startContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	id := mux.Vars(req)["name"]

	detachKeys := req.FormValue("detachKeys")

	if err := s.ContainerMgr.Start(ctx, id, detachKeys); err != nil {
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
	cis, err := s.ContainerMgr.List(ctx)
	if err != nil {
		return err
	}

	containerList := make([]types.Container, 0, len(cis))
	for _, ci := range cis {
		container := types.Container{
			ID:         ci.ID,
			Names:      []string{ci.Name},
			Status:     string(ci.State.Status),
			Image:      ci.Config.Image,
			Command:    strings.Join(ci.Config.Cmd, " "),
			Created:    ci.State.StartedAt,
			Labels:     ci.Config.Labels,
			HostConfig: ci.HostConfig,
		}
		containerList = append(containerList, container)
	}

	return EncodeResponse(rw, http.StatusOK, containerList)
}

func (s *Server) getContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]
	containerInfo, err := s.ContainerMgr.Get(name)
	if err != nil {
		return err
	}

	container := types.ContainerJSON{
		ID:      containerInfo.ID,
		Name:    containerInfo.Name,
		Image:   containerInfo.Config.Image,
		Created: containerInfo.State.StartedAt,
		State:   &types.ContainerState{Pid: containerInfo.State.Pid},
	}
	return EncodeResponse(rw, http.StatusOK, container)
}
