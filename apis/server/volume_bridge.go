package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/randomid"
	"github.com/alibaba/pouch/pkg/utils"
	volumetypes "github.com/alibaba/pouch/storage/volume/types"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s *Server) createVolume(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.VolumeCreateConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	logCreateOptions("volume", config)

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
	filter, err := filters.FromParam(req.FormValue("filters"))
	if err != nil {
		return err
	}

	volumes, err := s.VolumeMgr.List(ctx, filter)
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

func (s *Server) pruneVolume(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	var volumePruneRep types.VolumePruneResp

	filter, err := filters.FromParam(req.FormValue("filters"))
	if err != nil {
		return err
	}

	volumeList, err := s.VolumeMgr.List(ctx, filter)
	toDeleteFlag := map[string]bool{}
	if err != nil {
		return err
	}

	_, err = s.ContainerMgr.List(ctx, &mgr.ContainerListOption{
		All: true,
		FilterFunc: func(c *mgr.Container) bool {
			if len(c.Mounts) == 0 {
				return false
			}
			for _, v := range c.Mounts {
				toDeleteFlag[v.Name] = true
			}

			return true
		},
	})
	if err != nil {
		return err
	}

	for _, volume := range volumeList {
		if toDeleteFlag[volume.Name] == false {
			volumePruneRep.VolumesDeleted = append(volumePruneRep.VolumesDeleted, volume.Name)
			vSize, err := utils.DirectorySize(volume.Path())
			if err != nil {
				logrus.Warnf("could not determine size of volume: %v", volume.Name)
			}
			volumePruneRep.SpaceReclaimed += vSize
			s.VolumeMgr.Remove(ctx, volume.Name)
		}
	}

	return EncodeResponse(rw, http.StatusOK, volumePruneRep)
}
