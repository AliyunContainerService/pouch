package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/httputils"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// versionMatcher defines to parse version url path.
const versionMatcher = "/v{version:[0-9.]+}"

func initRoute(s *Server) http.Handler {
	r := mux.NewRouter()

	// system
	s.addRoute(r, http.MethodGet, "/_ping", s.ping)
	s.addRoute(r, http.MethodGet, "/info", s.info)
	s.addRoute(r, http.MethodGet, "/version", s.version)
	s.addRoute(r, http.MethodPost, "/auth", s.auth)

	// daemon, we still list this API into system manager.
	s.addRoute(r, http.MethodPost, "/daemon/update", s.updateDaemon)

	// container
	s.addRoute(r, http.MethodPost, "/containers/create", s.createContainer)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/start", s.startContainer)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/stop", s.stopContainer)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/attach", s.attachContainer)
	s.addRoute(r, http.MethodGet, "/containers/json", s.getContainers)
	s.addRoute(r, http.MethodGet, "/containers/{name:.*}/json", s.getContainer)
	s.addRoute(r, http.MethodDelete, "/containers/{name:.*}", s.removeContainers)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/exec", s.createContainerExec)
	s.addRoute(r, http.MethodGet, "/exec/{name:.*}/json", s.getExecInfo)
	s.addRoute(r, http.MethodPost, "/exec/{name:.*}/start", s.startContainerExec)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/rename", s.renameContainer)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/restart", s.restartContainer)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/pause", s.pauseContainer)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/unpause", s.unpauseContainer)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/update", s.updateContainer)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/upgrade", s.upgradeContainer)
	s.addRoute(r, http.MethodGet, "/containers/{name:.*}/top", s.topContainer)
	s.addRoute(r, http.MethodGet, "/containers/{name:.*}/logs", s.logsContainer)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/resize", s.resizeContainer)
	s.addRoute(r, http.MethodPost, "/containers/{name:.*}/restart", s.restartContainer)

	// image
	s.addRoute(r, http.MethodPost, "/images/create", s.pullImage)
	s.addRoute(r, http.MethodPost, "/images/search", s.searchImages)
	s.addRoute(r, http.MethodGet, "/images/json", s.listImages)
	s.addRoute(r, http.MethodDelete, "/images/{name:.*}", s.removeImage)
	s.addRoute(r, http.MethodGet, "/images/{name:.*}/json", s.getImage)

	// volume
	s.addRoute(r, http.MethodGet, "/volumes", s.listVolume)
	s.addRoute(r, http.MethodPost, "/volumes/create", s.createVolume)
	s.addRoute(r, http.MethodGet, "/volumes/{name:.*}", s.getVolume)
	s.addRoute(r, http.MethodDelete, "/volumes/{name:.*}", s.removeVolume)

	// network
	s.addRoute(r, http.MethodGet, "/networks", s.listNetwork)
	s.addRoute(r, http.MethodPost, "/networks/create", s.createNetwork)
	s.addRoute(r, http.MethodGet, "/networks/{name:.*}", s.getNetwork)
	s.addRoute(r, http.MethodDelete, "/networks/{name:.*}", s.deleteNetwork)
	s.addRoute(r, http.MethodPost, "/networks/{name:.*}/disconnect", s.disconnectNetwork)

	// metrics
	r.Path(versionMatcher + "/metrics").Methods(http.MethodGet).Handler(prometheus.Handler())
	r.Path("/metrics").Methods(http.MethodGet).Handler(prometheus.Handler())

	if s.Config.Debug || s.Config.EnableProfiler {
		profilerSetup(r)
	}
	return r
}

func (s *Server) addRoute(r *mux.Router, mothod string, path string, f func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error) {
	r.Path(versionMatcher + path).Methods(mothod).Handler(filter(f, s))
	r.Path(path).Methods(mothod).Handler(filter(f, s))
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

func filter(handler handler, s *Server) http.HandlerFunc {
	pctx := context.Background()

	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithCancel(pctx)
		defer cancel()

		s.lock.RLock()
		if len(s.ManagerWhiteList) > 0 && req.TLS != nil && len(req.TLS.PeerCertificates) > 0 {
			if _, isManager := s.ManagerWhiteList[req.TLS.PeerCertificates[0].Subject.CommonName]; !isManager {
				s.lock.RUnlock()
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("tls verified error."))
				return
			}
		}
		s.lock.RUnlock()

		t := time.Now()
		clientInfo := req.RemoteAddr
		defer func() {
			d := time.Since(t) / (time.Millisecond)
			// If elapse time of handler >= 500ms, log request.
			if d >= 500 {
				logrus.Infof("End of Calling %s %s, costs %d ms. client %s", req.Method, req.URL.Path, d, clientInfo)
			}
		}()

		if req.TLS != nil && len(req.TLS.PeerCertificates) > 0 {
			issuer := req.TLS.PeerCertificates[0].Issuer.CommonName
			clientName := req.TLS.PeerCertificates[0].Subject.CommonName
			clientInfo = fmt.Sprintf("%s %s %s", clientInfo, issuer, clientName)
		}
		if req.Method != http.MethodGet {
			logrus.Infof("Calling %s %s, client %s", req.Method, req.URL.RequestURI(), clientInfo)
		} else {
			logrus.Debugf("Calling %s %s, client %s", req.Method, req.URL.RequestURI(), clientInfo)
		}

		// Start to handle request.
		err := handler(ctx, w, req)
		if err == nil {
			return
		}
		// Handle error if request handling fails.
		HandleErrorResponse(w, err)
	}
}

// EncodeResponse encodes response in json.
func EncodeResponse(rw http.ResponseWriter, statusCode int, data interface{}) error {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	return json.NewEncoder(rw).Encode(data)
}

// HandleErrorResponse handles err from daemon side and constructs response for client side.
func HandleErrorResponse(w http.ResponseWriter, err error) {
	var (
		code   int
		errMsg string
	)

	// By default, daemon side returns code 500 if error happens.
	code = http.StatusInternalServerError
	errMsg = err.Error()

	httpErr, ok := err.(httputils.HTTPError)
	if ok {
		code = httpErr.Code()
	} else if errtypes.IsNotfound(err) {
		code = http.StatusNotFound
	} else if errtypes.IsInvalidParam(err) {
		code = http.StatusBadRequest
	} else if errtypes.IsAlreadyExisted(err) {
		code = http.StatusConflict
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	resp := types.Error{
		Message: errMsg,
	}
	enc.Encode(resp)
}
