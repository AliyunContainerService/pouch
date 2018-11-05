package server

import (
	"context"
	"net/http"
)

func (s *Server) criExec(context context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	if s.StreamRouter == nil {
		return EncodeResponse(rw, http.StatusNotImplemented, nil)
	}
	s.StreamRouter.ServeExec(rw, req)
	return nil
}

func (s *Server) criAttach(context context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	if s.StreamRouter == nil {
		return EncodeResponse(rw, http.StatusNotImplemented, nil)
	}
	s.StreamRouter.ServeAttach(rw, req)
	return nil
}

func (s *Server) criPortForward(context context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	if s.StreamRouter == nil {
		return EncodeResponse(rw, http.StatusNotImplemented, nil)
	}
	s.StreamRouter.ServePortForward(rw, req)
	return nil
}
