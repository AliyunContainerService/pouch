package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/pkg/utils/metrics"

	"github.com/docker/docker/pkg/ioutils"
	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *Server) ping(context context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte{'O', 'K'})
	return
}

func (s *Server) info(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	info, err := s.SystemMgr.Info()
	if err != nil {
		return err
	}
	return EncodeResponse(rw, http.StatusOK, info)
}

func (s *Server) version(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	version, err := s.SystemMgr.Version()
	if err != nil {
		return err
	}
	return EncodeResponse(rw, http.StatusOK, version)
}

func (s *Server) updateDaemon(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	cfg := &types.DaemonUpdateConfig{}

	// decode request body
	if err := json.NewDecoder(req.Body).Decode(cfg); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := cfg.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	return s.SystemMgr.UpdateDaemon(cfg)
}

func (s *Server) auth(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	auth := types.AuthConfig{}

	// decode request body
	if err := json.NewDecoder(req.Body).Decode(&auth); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := auth.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	token, err := s.SystemMgr.Auth(&auth)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	authResp := types.AuthResponse{
		Status:        "Login Succeeded",
		IdentityToken: token,
	}
	return EncodeResponse(rw, http.StatusOK, authResp)
}

func (s *Server) events(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	rw.Header().Set("Content-Type", "application/json")
	output := ioutils.NewWriteFlusher(rw)
	defer output.Close()
	output.Flush()
	enc := json.NewEncoder(output)

	// parse the since and until parameters
	since, err := eventTime(req.FormValue("since"))
	if err != nil {
		return err
	}
	until, err := eventTime(req.FormValue("until"))
	if err != nil {
		return err
	}

	var (
		timeout        <-chan time.Time
		onlyPastEvents bool
	)
	if !until.IsZero() {
		if until.Before(since) {
			return fmt.Errorf("until time (%s) cannot be after since (%s)", req.FormValue("until"), req.FormValue("since"))
		}

		now := time.Now()
		onlyPastEvents = until.Before(now)
		if !onlyPastEvents {
			dur := until.Sub(now)
			timeout = time.NewTimer(dur).C
		}
	}

	ef, err := filters.FromParam(req.FormValue("filters"))
	if err != nil {
		return err
	}

	// send past events
	buffered, eventq, errq := s.SystemMgr.SubscribeToEvents(ctx, since, until, ef)
	for _, ev := range buffered {
		if err := enc.Encode(ev); err != nil {
			return err
		}
	}

	// if until time is before now(), we only send past events
	if onlyPastEvents {
		return nil
	}

	// start subscribe new pouchd events
	for {
		select {
		case ev := <-eventq:
			if err := enc.Encode(ev); err != nil {
				logrus.Errorf("encode events got an error: %v", err)
				return err
			}
		case err := <-errq:
			if err != nil {
				return errors.Wrapf(err, "subscribe failed")
			}
			return nil
		case <-timeout:
			return nil
		case <-ctx.Done():
			logrus.Debug("client context is cancelled, stop sending events")
			return nil
		}
	}
}

func (s *Server) metrics(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	metrics.GetPrometheusHandler().ServeHTTP(rw, req)
	return nil
}

func eventTime(formTime string) (time.Time, error) {
	t, tNano, err := utils.ParseTimestamp(formTime, -1)
	if err != nil {
		return time.Time{}, err
	}
	if t == -1 {
		return time.Time{}, nil
	}
	return time.Unix(t, tNano), nil
}
