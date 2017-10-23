package server

import (
	"context"
	"encoding/json"
	"net/http"
)

func (s *Server) ping(context context.Context, resp http.ResponseWriter, req *http.Request) (err error) {
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte{'O', 'K'})
	return
}

func (s *Server) info(ctx context.Context, resp http.ResponseWriter, req *http.Request) (err error) {
	info, err := s.SystemMgr.Info()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	resp.WriteHeader(http.StatusOK)
	return json.NewEncoder(resp).Encode(info)
}

func (s *Server) version(ctx context.Context, resp http.ResponseWriter, req *http.Request) (err error) {
	info, err := s.SystemMgr.Version()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	resp.WriteHeader(http.StatusOK)
	return json.NewEncoder(resp).Encode(info)
}
