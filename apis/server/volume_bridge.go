package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/randomid"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
)

func (s *Server) createVolume(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.VolumeCreateConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := config.Name
	driver := config.Driver
	options := config.DriverOpts
	labels := config.Labels

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
		Labels: config.Labels,
	}
	return EncodeResponse(rw, http.StatusCreated, volume)
}

func (s *Server) getVolume(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	name := mux.Vars(req)["name"]
	volume, err := s.VolumeMgr.Get(ctx, name)
	if err != nil {
		return err
	}
	return EncodeResponse(rw, http.StatusOK, volume)
}

func (s *Server) removeVolume(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	name := mux.Vars(req)["name"]

	if err := s.VolumeMgr.Remove(ctx, name); err != nil {
		return err
	}
	rw.WriteHeader(http.StatusNoContent)
	return nil
}
