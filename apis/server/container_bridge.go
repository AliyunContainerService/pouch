package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/gorilla/mux"
)

func (s *Server) createContainer(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	var config types.ContainerConfigWrapper

	if err := json.NewDecoder(req.Body).Decode(&config); err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return err
	}

	name := req.FormValue("name")

	ret, err := s.ContainerMgr.Create(ctx, name, &config)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
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
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}

	resp.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) stopContainer(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	t, err := strconv.Atoi(req.FormValue("t"))
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return err
	}

	name := mux.Vars(req)["name"]

	if err := s.ContainerMgr.Stop(ctx, name, time.Duration(t)*time.Second); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}

	resp.WriteHeader(http.StatusOK)
	return nil
}
