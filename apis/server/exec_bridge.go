package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/streams"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s *Server) createContainerExec(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.ExecCreateConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	name := mux.Vars(req)["name"]

	logCreateOptions("container exec for "+name, config)

	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	id, err := s.ContainerMgr.CreateExec(ctx, name, config)
	if err != nil {
		return err
	}

	execCreateResp := &types.ExecCreateResp{
		ID: id,
	}

	return EncodeResponse(rw, http.StatusCreated, execCreateResp)
}

func (s *Server) startContainerExec(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.ExecStartConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	name := mux.Vars(req)["name"]
	_, upgrade := req.Header["Upgrade"]

	ba, _ := json.Marshal(config)
	logrus.Infof("start exec %s, upgrade: %v, body: %s", name, upgrade, string(ba))

	if err := s.ContainerMgr.CheckExecExist(ctx, name); err != nil {
		return err
	}

	var (
		err     error
		closeFn func() error
		attach  = new(streams.AttachConfig)
		stdin   io.ReadCloser
		stdout  io.Writer
	)

	// TODO(huamin.thm): support detach exec process through http post method
	if !config.Detach {
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

		attach.UseStdin, attach.Stdin = true, stdin
		attach.Terminal = config.Tty

		if config.Tty {
			attach.UseStdout, attach.Stdout = true, stdout
		} else {
			attach.UseStdout, attach.Stdout = true, stdcopy.NewStdWriter(stdout, stdcopy.Stdout)
			attach.UseStderr, attach.Stderr = true, stdcopy.NewStdWriter(stdout, stdcopy.Stderr)
		}
	}

	if err := s.ContainerMgr.StartExec(ctx, name, attach); err != nil {
		if config.Detach {
			return err
		}
		attach.Stdout.Write([]byte(err.Error() + "\r\n"))
	}
	return nil
}

func (s *Server) getExecInfo(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]
	execInfo, err := s.ContainerMgr.InspectExec(ctx, name)
	if err != nil {
		return err
	}
	return EncodeResponse(rw, http.StatusOK, execInfo)
}

func (s *Server) resizeExec(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
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

	if err := s.ContainerMgr.ResizeExec(ctx, name, opts); err != nil {
		return err
	}

	rw.WriteHeader(http.StatusOK)
	return nil

}

func openHijackConnection(rw http.ResponseWriter) (io.ReadCloser, io.Writer, func() error, error) {
	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		return nil, nil, nil, fmt.Errorf("not a hijack connection")
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, nil, err
	}

	// set raw mode
	conn.Write([]byte{})
	return conn, conn, func() error {
		return conn.Close()
	}, nil
}
