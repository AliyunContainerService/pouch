package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/randomid"

	"github.com/gorilla/mux"
)

func (s *Server) createVolume(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	var volumeCreateReq types.VolumeCreateRequest

	if err := json.NewDecoder(req.Body).Decode(&volumeCreateReq); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := volumeCreateReq.Name
	driver := volumeCreateReq.Driver
	options := volumeCreateReq.DriverOpts
	labels := volumeCreateReq.Labels

	if name == "" {
		name = randomid.Generate()
	}

	if driver == "" {
		driver = "local"
	}

	if err := s.VolumeMgr.Create(ctx, name, driver, options, labels); err != nil {
		return err
	}

	volume := types.VolumeInfo{
		Name:   name,
		Driver: driver,
		Labels: volumeCreateReq.Labels,
	}
	resp.WriteHeader(http.StatusCreated)
	return json.NewEncoder(resp).Encode(volume)
}

func (s *Server) removeVolume(ctx context.Context, resp http.ResponseWriter, req *http.Request) (err error) {
	name := mux.Vars(req)["name"]

	if err := s.VolumeMgr.Remove(ctx, name); err != nil {
		return err
	}
	resp.WriteHeader(http.StatusOK)
	return nil
}
