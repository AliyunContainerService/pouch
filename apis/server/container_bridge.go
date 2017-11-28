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
	"github.com/alibaba/pouch/pkg/httputils"

	"github.com/gorilla/mux"
)

// TODO
func (s *Server) removeContainers(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	return nil
}

func (s *Server) createContainerExec(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
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

	resp.WriteHeader(http.StatusCreated)
	return json.NewEncoder(resp).Encode(&types.ExecCreateResponse{
		ID: id,
	})
}

func (s *Server) startContainerExec(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	config := &types.ExecStartConfig{}

	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return err
	}

	var attach *types.AttachConfig

	if !config.Detach {
		hijacker, ok := resp.(http.Hijacker)
		if !ok {
			return fmt.Errorf("not a hijack connection, container: %s", name)
		}

		attach = &types.AttachConfig{
			Hijack:  hijacker,
			Stdin:   config.Tty,
			Stdout:  true,
			Stderr:  true,
			Upgrade: true,
		}
	}

	return s.ContainerMgr.StartExec(ctx, name, config, attach)
}

func (s *Server) createContainer(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	var config types.ContainerConfigWrapper

	if err := json.NewDecoder(req.Body).Decode(&config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := req.FormValue("name")

	ret, err := s.ContainerMgr.Create(ctx, name, &config)
	if err != nil {
		return err
	}

	resp.WriteHeader(http.StatusCreated)
	return json.NewEncoder(resp).Encode(ret)
}

func (s *Server) startContainer(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]
	config := types.ContainerStartConfig{
		ID:         name,
		DetachKeys: req.FormValue("detachKeys"),
	}

	if err := s.ContainerMgr.Start(ctx, config); err != nil {
		return err
	}

	resp.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) stopContainer(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
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

	if err = s.ContainerMgr.Stop(ctx, name, time.Duration(t)*time.Second); err != nil {
		return err
	}

	resp.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) attachContainer(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	_, upgrade := req.Header["Upgrade"]

	hijacker, ok := resp.(http.Hijacker)
	if !ok {
		return fmt.Errorf("not a hijack connection, container: %s", name)
	}

	attach := &types.AttachConfig{
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

func (s *Server) getContainers(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	cis, err := s.ContainerMgr.List(ctx)
	if err != nil {
		return err
	}

	cs := make([]types.Container, 0, len(cis))
	for _, ci := range cis {
		c := types.Container{
			ID:      ci.ID,
			Names:   []string{ci.Name},
			Status:  ci.Status.String(),
			Image:   ci.Config.Image,
			Command: strings.Join(ci.Config.Cmd, " "),
			Created: ci.StartedAt.UnixNano(),
			Labels:  ci.Config.Labels,
		}
		cs = append(cs, c)
	}

	resp.WriteHeader(http.StatusOK)
	return json.NewEncoder(resp).Encode(cs)
}
