package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/httputils"
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

func (s *Server) auth(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	auth := types.AuthConfig{}
	if err := json.NewDecoder(req.Body).Decode(&auth); err != nil {
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
