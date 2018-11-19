package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/metrics"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/streams"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/pkg/utils/filters"
	util_metrics "github.com/alibaba/pouch/pkg/utils/metrics"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *Server) createContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	label := util_metrics.ActionCreateLabel
	metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
	defer func(start time.Time) {
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	config := &types.ContainerCreateConfig{}
	reader := req.Body
	// decode request body
	if err := json.NewDecoder(reader).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	logCreateOptions("container", config)

	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := req.FormValue("name")
	//consider set specific id by url params
	specificID := req.FormValue("specificId")
	if specificID != "" {
		config.SpecificID = specificID
	}

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

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return EncodeResponse(rw, http.StatusCreated, container)
}

func (s *Server) getContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	c, err := s.ContainerMgr.Get(ctx, name)
	if err != nil {
		return err
	}

	mounts := []types.MountPoint{}
	for _, mp := range c.Mounts {
		mounts = append(mounts, *mp)
	}

	container := types.ContainerJSON{
		ID:           c.ID,
		Name:         c.Name,
		Image:        c.Image,
		Created:      c.Created,
		State:        c.State,
		Config:       c.Config,
		HostConfig:   c.HostConfig,
		LogPath:      c.LogPath,
		Snapshotter:  c.Snapshotter,
		RestartCount: c.RestartCount,
		GraphDriver: &types.GraphDriverData{
			Name: c.Snapshotter.Name,
			Data: c.Snapshotter.Data,
		},
		Mounts:          mounts,
		NetworkSettings: c.NetworkSettings,
	}

	return EncodeResponse(rw, http.StatusOK, container)
}

func (s *Server) getContainers(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	option := &mgr.ContainerListOption{
		All: httputils.BoolValue(req, "all"),
	}

	filters, err := filters.FromURLParam(req.FormValue("filters"))
	if err != nil {
		return err
	}
	option.Filter = filters

	cons, err := s.ContainerMgr.List(ctx, option)
	if err != nil {
		return err
	}

	containerList := make([]types.Container, 0, len(cons))

	for _, c := range cons {
		status, err := c.FormatStatus()
		if err != nil {
			return err
		}

		t, err := time.Parse(utils.TimeLayout, c.Created)
		if err != nil {
			return err
		}

		var netSettings *types.ContainerNetworkSettings
		if c.NetworkSettings != nil {
			netSettings = &types.ContainerNetworkSettings{
				Networks: c.NetworkSettings.Networks,
			}
		}

		singleCon := types.Container{
			ID:              c.ID,
			Names:           []string{c.Name},
			Image:           c.Config.Image,
			ImageID:         c.Image,
			Command:         strings.Join(c.Config.Cmd, " "),
			Status:          status,
			Created:         t.UnixNano(),
			Labels:          c.Config.Labels,
			HostConfig:      c.HostConfig,
			NetworkSettings: netSettings,
		}

		containerList = append(containerList, singleCon)
	}
	return EncodeResponse(rw, http.StatusOK, containerList)
}

func (s *Server) startContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	label := util_metrics.ActionStartLabel
	metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
	defer func(start time.Time) {
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	name := mux.Vars(req)["name"]

	options := &types.ContainerStartOptions{
		DetachKeys:    req.FormValue("detachKeys"),
		CheckpointID:  req.FormValue("checkpoint"),
		CheckpointDir: req.FormValue("checkpoint-dir"),
	}

	if err := s.ContainerMgr.Start(ctx, name, options); err != nil {
		return err
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) restartContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	var (
		t   int
		err error
	)
	label := util_metrics.ActionRestartLabel
	metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
	defer func(start time.Time) {
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	if v := req.FormValue("t"); v != "" {
		if t, err = strconv.Atoi(v); err != nil {
			return httputils.NewHTTPError(err, http.StatusBadRequest)
		}
	}

	name := mux.Vars(req)["name"]

	if err = s.ContainerMgr.Restart(ctx, name, int64(t)); err != nil {
		return err
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) stopContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	var (
		t   int
		err error
	)

	label := util_metrics.ActionStopLabel
	metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
	defer func(start time.Time) {
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	if v := req.FormValue("t"); v != "" {
		if t, err = strconv.Atoi(v); err != nil {
			return httputils.NewHTTPError(err, http.StatusBadRequest)
		}
	}

	name := mux.Vars(req)["name"]

	if err = s.ContainerMgr.Stop(ctx, name, int64(t)); err != nil {
		return err
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

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

func (s *Server) renameContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	label := util_metrics.ActionRenameLabel
	metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
	defer func(start time.Time) {
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	oldName := mux.Vars(req)["name"]
	newName := req.FormValue("name")

	if err := s.ContainerMgr.Rename(ctx, oldName, newName); err != nil {
		return err
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) attachContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]
	_, upgrade := req.Header["Upgrade"]

	var (
		err     error
		closeFn func() error
		attach  = new(streams.AttachConfig)
		stdin   io.ReadCloser
		stdout  io.Writer
	)

	stdin, stdout, closeFn, err = openHijackConnection(rw)
	if err != nil {
		return err
	}

	// close hijack stream
	defer closeFn()

	if upgrade {
		fmt.Fprintf(stdout, "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n")
	} else {
		fmt.Fprintf(stdout, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n\r\n")
	}

	attach.UseStdin = httputils.BoolValue(req, "stdin")
	attach.Stdin = stdin
	attach.UseStdout = true
	attach.Stdout = stdout
	attach.UseStderr = true
	attach.Stderr = stdout

	if err := s.ContainerMgr.AttachContainerIO(ctx, name, attach); err != nil {
		stdout.Write([]byte(err.Error() + "\r\n"))
	}
	return nil
}

func (s *Server) updateContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	label := util_metrics.ActionUpdateLabel
	metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
	defer func(start time.Time) {
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	config := &types.UpdateConfig{}

	// set pre update hook plugin
	reader := req.Body
	if s.ContainerPlugin != nil {
		var err error
		logrus.Infof("invoke container pre-update hook in plugin")
		if reader, err = s.ContainerPlugin.PreUpdate(req.Body); err != nil {
			return errors.Wrapf(err, "failed to execute pre-create plugin point")
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

	name := mux.Vars(req)["name"]

	if err := s.ContainerMgr.Update(ctx, name, config); err != nil {
		return httputils.NewHTTPError(err, http.StatusInternalServerError)
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	rw.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) upgradeContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	label := util_metrics.ActionUpgradeLabel
	metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
	defer func(start time.Time) {
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

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

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	rw.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) topContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	query := req.URL.Query()
	procList, err := s.ContainerMgr.Top(ctx, name, query.Get("ps_args"))
	if err != nil {
		return err
	}

	return EncodeResponse(rw, http.StatusOK, procList)
}

func (s *Server) logsContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	opts := &types.ContainerLogsOptions{
		ShowStdout: httputils.BoolValue(req, "stdout"),
		ShowStderr: httputils.BoolValue(req, "stderr"),

		Tail:       req.Form.Get("tail"),
		Since:      req.Form.Get("since"),
		Until:      req.Form.Get("until"),
		Follow:     httputils.BoolValue(req, "follow"),
		Timestamps: httputils.BoolValue(req, "timestamps"),
		Details:    httputils.BoolValue(req, "details"),
	}

	name := mux.Vars(req)["name"]
	msgCh, tty, err := s.ContainerMgr.Logs(ctx, name, opts)
	if err != nil {
		return err
	}
	writeLogStream(ctx, rw, tty, opts, msgCh)
	return nil
}

func (s *Server) statsContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	stream := httputils.BoolValue(req, "stream")

	if !stream {
		rw.Header().Set("Content-Type", "application/json")
	}

	config := &mgr.ContainerStatsConfig{
		Stream:    stream,
		OutStream: rw,
	}

	return s.ContainerMgr.StreamStats(ctx, name, config)
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

func (s *Server) removeContainers(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	label := util_metrics.ActionDeleteLabel
	metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
	defer func(start time.Time) {
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	name := mux.Vars(req)["name"]

	option := &types.ContainerRemoveOptions{
		Force:   httputils.BoolValue(req, "force"),
		Volumes: httputils.BoolValue(req, "v"),
		// TODO: Link will be supported in the future.
		Link: httputils.BoolValue(req, "link"),
	}

	if err := s.ContainerMgr.Remove(ctx, name, option); err != nil {
		return err
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) waitContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	waitStatus, err := s.ContainerMgr.Wait(ctx, name)

	if err != nil {
		return err
	}

	return EncodeResponse(rw, http.StatusOK, &waitStatus)
}

func (s *Server) createContainerCheckpoint(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	options := &types.CheckpointCreateOptions{}
	if err := json.NewDecoder(req.Body).Decode(options); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	// ensure CheckpointID should not be empty
	if options.CheckpointID == "" {
		return httputils.NewHTTPError(fmt.Errorf("checkpoint id should not be empty"), http.StatusBadRequest)
	}

	if err := s.ContainerMgr.CreateCheckpoint(ctx, name, options); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusCreated)
	return nil
}

func (s *Server) listContainerCheckpoint(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	options := &types.CheckpointListOptions{
		CheckpointDir: req.FormValue("dir"),
	}

	list, err := s.ContainerMgr.ListCheckpoint(ctx, name, options)
	if err != nil {
		return err
	}

	return EncodeResponse(rw, http.StatusOK, list)
}

func (s *Server) deleteContainerCheckpoint(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	options := &types.CheckpointDeleteOptions{
		CheckpointID:  mux.Vars(req)["id"],
		CheckpointDir: req.FormValue("dir"),
	}

	// ensure CheckpointID should not be empty
	if options.CheckpointID == "" {
		return httputils.NewHTTPError(fmt.Errorf("checkpoint id should not be empty"), http.StatusBadRequest)
	}

	if err := s.ContainerMgr.DeleteCheckpoint(ctx, name, options); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) commitContainer(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	options := &types.ContainerCommitOptions{
		Repository: req.FormValue("repo"),
		Tag:        req.FormValue("tag"),
		Author:     req.FormValue("author"),
		Comment:    req.FormValue("comment"),
	}

	id, err := s.ContainerMgr.Commit(ctx, req.FormValue("container"), options)
	if err != nil {
		return err
	}

	return EncodeResponse(rw, http.StatusCreated, id)
}
