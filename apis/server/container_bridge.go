package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
	"github.com/gorilla/mux"
)

func (s *Server) createContainer(ctx context.Context, resp http.ResponseWriter, req *http.Request) (err error) {
	var config types.ContainerConfigWrapper
	name := req.FormValue("name")

	err = json.NewDecoder(req.Body).Decode(&config)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return err
	}

	ret, err := s.ContainerMgr.Create(ctx, name, &config)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	//resp.WriteHeader(http.StatusCreated)
	return json.NewEncoder(resp).Encode(ret)
}

func (s *Server) startContainer(ctx context.Context, resp http.ResponseWriter, req *http.Request) (err error) {
	name := mux.Vars(req)["name"]
	config := types.ContainerStartConfig{
		ID:         name,
		DetachKeys: req.FormValue("detachKeys"),
	}
	err = s.ContainerMgr.Start(ctx, config)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	resp.WriteHeader(http.StatusOK)
	return nil
}
