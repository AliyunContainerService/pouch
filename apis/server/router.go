package server

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func initRoute(s *Server) http.Handler {
	r := mux.NewRouter()
	// system
	r.Path("/_ping").Methods(http.MethodGet).Handler(s.filter(s.ping))
	r.Path("/info").Methods(http.MethodGet).Handler(s.filter(s.info))
	r.Path("/version").Methods(http.MethodGet).Handler(s.filter(s.version))

	// container
	r.Path("/containers/create").Methods(http.MethodPost).Handler(s.filter(s.createContainer))
	r.Path("/containers/{name:.*}/start").Methods(http.MethodPost).Handler(s.filter(s.startContainer))

	// image
	r.Path("/images/create").Methods(http.MethodPost).Handler(s.filter(s.pullImage))
	r.Path("/images/search").Methods(http.MethodGet).Handler(s.filter(s.searchImages))
	r.Path("/images/json").Methods(http.MethodGet).Handler(s.filter(s.listImages))

	// volume
	r.Path("/volumes/create").Methods(http.MethodPost).Handler(s.filter(s.createVolume))
	r.Path("/volumes/{name:.*}").Methods(http.MethodDelete).Handler(s.filter(s.removeVolume))

	if s.Config.Debug {
		profilerSetup(r)
	}
	return r
}

func profilerSetup(mainRouter *mux.Router) {
	var r = mainRouter.PathPrefix("/debug/").Subrouter()
	r.HandleFunc("/pprof/", pprof.Index)
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/profile", pprof.Profile)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)
	r.HandleFunc("/pprof/block", pprof.Handler("block").ServeHTTP)
	r.HandleFunc("/pprof/heap", pprof.Handler("heap").ServeHTTP)
	r.HandleFunc("/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	r.HandleFunc("/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
}

type handler func(context.Context, http.ResponseWriter, *http.Request) error

func (s *Server) filter(handler handler) http.HandlerFunc {
	pctx := context.Background()

	return func(resp http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithCancel(pctx)
		defer cancel()

		// TODO: add TLS verify

		if req.Method != http.MethodGet {
			logrus.Infof("Calling %s %s", req.Method, req.URL.RequestURI())
		}

		if err := handler(ctx, resp, req); err != nil {
			logrus.Errorf("invoke %s error %v", req.URL.RequestURI(), err)
			resp.Write([]byte(err.Error()))
		}
	}
}
