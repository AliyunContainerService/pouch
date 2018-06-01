package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/randomid"
	volumetypes "github.com/alibaba/pouch/storage/volume/types"

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
		driver = volumetypes.DefaultBackend
	}

	volume, err := s.VolumeMgr.Create(ctx, name, driver, options, labels)
	if err != nil {
		return err
	}

	status := map[string]interface{}{}
	for k, v := range volume.Options() {
		if k != "" && v != "" {
			status[k] = v
		}
	}
	status["size"] = volume.Size()

	respVolume := types.VolumeInfo{
		Name:       name,
		Driver:     driver,
		Labels:     config.Labels,
		Mountpoint: volume.Path(),
		Status:     status,
		CreatedAt:  volume.CreationTimestamp.Format("2006-1-2 15:04:05"),
	}

	return EncodeResponse(rw, http.StatusCreated, respVolume)
}

func (s *Server) getVolume(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	name := mux.Vars(req)["name"]
	volume, err := s.VolumeMgr.Get(ctx, name)
	if err != nil {
		return err
	}

	status := map[string]interface{}{}
	for k, v := range volume.Options() {
		if k != "" && v != "" {
			status[k] = v
		}
	}
	status["size"] = volume.Size()

	respVolume := types.VolumeInfo{
		Name:       volume.Name,
		Driver:     volume.Driver(),
		Mountpoint: volume.Path(),
		CreatedAt:  volume.CreateTime(),
		Labels:     volume.Labels,
		Status:     status,
	}

	return EncodeResponse(rw, http.StatusOK, respVolume)
}

func (s *Server) listVolume(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	volumes, err := s.VolumeMgr.List(ctx, map[string]string{})
	if err != nil {
		return err
	}

	respVolumes := types.VolumeListResp{Volumes: []*types.VolumeInfo{}, Warnings: nil}
	for _, volume := range volumes {
		status := map[string]interface{}{}
		for k, v := range volume.Options() {
			if k != "" && v != "" {
				status[k] = v
			}
		}
		status["size"] = volume.Size()

		respVolume := &types.VolumeInfo{
			Name:       volume.Name,
			Driver:     volume.Driver(),
			Mountpoint: volume.Path(),
			CreatedAt:  volume.CreateTime(),
			Labels:     volume.Labels,
			Status:     status,
		}
		respVolumes.Volumes = append(respVolumes.Volumes, respVolume)
	}
	return EncodeResponse(rw, http.StatusOK, respVolumes)
}

func (s *Server) removeVolume(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	if err := s.VolumeMgr.Remove(ctx, name); err != nil {
		return err
	}
	rw.WriteHeader(http.StatusNoContent)
	return nil
}
