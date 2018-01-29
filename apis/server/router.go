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

func initRoute(s *Server) http.Handler {
	r := mux.NewRouter()
	// system
	r.Path("/_ping").Methods(http.MethodGet).Handler(s.filter(s.ping))
	r.Path("/info").Methods(http.MethodGet).Handler(s.filter(s.info))
	r.Path("/version").Methods(http.MethodGet).Handler(s.filter(s.version))

	// container
	r.Path("/containers/create").Methods(http.MethodPost).Handler(s.filter(s.createContainer))
	r.Path("/containers/{name:.*}/start").Methods(http.MethodPost).Handler(s.filter(s.startContainer))
	r.Path("/containers/{name:.*}/stop").Methods(http.MethodPost).Handler(s.filter(s.stopContainer))
	r.Path("/containers/{name:.*}/attach").Methods(http.MethodPost).Handler(s.filter(s.attachContainer))
	r.Path("/containers/json").Methods(http.MethodGet).Handler(s.filter(s.getContainers))
	r.Path("/containers/{name:.*}/json").Methods(http.MethodGet).Handler(s.filter(s.getContainer))
	r.Path("/containers/{name:.*}").Methods(http.MethodDelete).Handler(s.filter(s.removeContainers))
	r.Path("/containers/{name:.*}/exec").Methods(http.MethodPost).Handler(s.filter(s.createContainerExec))
	r.Path("/exec/{name:.*}/start").Methods(http.MethodPost).Handler(s.filter(s.startContainerExec))
	r.Path("/containers/{id:.*}/rename").Methods(http.MethodPost).Handler(s.filter(s.renameContainer))
	r.Path("/containers/{name:.*}/pause").Methods(http.MethodPost).Handler(s.filter(s.pauseContainer))
	r.Path("/containers/{name:.*}/unpause").Methods(http.MethodPost).Handler(s.filter(s.unpauseContainer))

	// image
	r.Path("/images/create").Methods(http.MethodPost).Handler(s.filter(s.pullImage))
	r.Path("/images/search").Methods(http.MethodGet).Handler(s.filter(s.searchImages))
	r.Path("/images/json").Methods(http.MethodGet).Handler(s.filter(s.listImages))
	r.Path("/images/{name:.*}").Methods(http.MethodDelete).Handler(s.filter(s.removeImage))
	r.Path("/images/{name:.*}/json").Methods(http.MethodGet).Handler(s.filter(s.getImage))

	// volume
	r.Path("/volumes").Methods(http.MethodGet).Handler(s.filter(s.listVolume))
	r.Path("/volumes/create").Methods(http.MethodPost).Handler(s.filter(s.createVolume))
	r.Path("/volumes/{name:.*}").Methods(http.MethodGet).Handler(s.filter(s.getVolume))
	r.Path("/volumes/{name:.*}").Methods(http.MethodDelete).Handler(s.filter(s.removeVolume))

	// metrics
	r.Path("/metrics").Methods(http.MethodGet).Handler(prometheus.Handler())

	// network
	r.Path("/networks").Methods(http.MethodGet).Handler(s.filter(s.listNetwork))
	r.Path("/networks/create").Methods(http.MethodPost).Handler(s.filter(s.createNetwork))
	r.Path("/networks/{name:.*}").Methods(http.MethodGet).Handler(s.filter(s.getNetwork))
	r.Path("/networks/{name:.*}").Methods(http.MethodDelete).Handler(s.filter(s.deleteNetwork))

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

	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithCancel(pctx)
		defer cancel()

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
