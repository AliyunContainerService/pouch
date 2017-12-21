package server

import (
	"context"
	"net/http"
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
