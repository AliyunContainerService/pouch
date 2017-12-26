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

func (s *Server) createVolume(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	var volumeCreateConfig types.VolumeCreateConfig

	if err := json.NewDecoder(req.Body).Decode(&volumeCreateConfig); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := volumeCreateConfig.Name
	driver := volumeCreateConfig.Driver
	options := volumeCreateConfig.DriverOpts
	labels := volumeCreateConfig.Labels

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
		Labels: volumeCreateConfig.Labels,
	}
	return EncodeResponse(rw, http.StatusCreated, volume)
}

func (s *Server) removeVolume(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	name := mux.Vars(req)["name"]

	if err := s.VolumeMgr.Remove(ctx, name); err != nil {
		return err
	}
	rw.WriteHeader(http.StatusOK)
	return nil
}
