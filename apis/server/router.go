package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	serverTypes "github.com/alibaba/pouch/apis/server/types"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// versionMatcher defines to parse version url path.
const versionMatcher = "/v{version:[0-9.]+}"

func initRoute(s *Server) *mux.Router {
	r := mux.NewRouter()

	handlers := []*serverTypes.HandlerSpec{
		// system
		{Method: http.MethodGet, Path: "/_ping", HandlerFunc: s.ping},
		{Method: http.MethodGet, Path: "/info", HandlerFunc: s.info},
		{Method: http.MethodGet, Path: "/version", HandlerFunc: s.version},
		{Method: http.MethodPost, Path: "/auth", HandlerFunc: s.auth},
		{Method: http.MethodGet, Path: "/events", HandlerFunc: withCancelHandler(s.events)},

		// daemon, we still list this API into system manager.
		{Method: http.MethodPost, Path: "/daemon/update", HandlerFunc: s.updateDaemon},

		// container
		{Method: http.MethodPost, Path: "/containers/{name:.*}/checkpoints", HandlerFunc: withCancelHandler(s.createContainerCheckpoint)},
		{Method: http.MethodGet, Path: "/containers/{name:.*}/checkpoints", HandlerFunc: withCancelHandler(s.listContainerCheckpoint)},
		{Method: http.MethodDelete, Path: "/containers/{name}/checkpoints/{id}", HandlerFunc: withCancelHandler(s.deleteContainerCheckpoint)},
		{Method: http.MethodPost, Path: "/containers/create", HandlerFunc: s.createContainer},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/start", HandlerFunc: s.startContainer},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/stop", HandlerFunc: s.stopContainer},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/attach", HandlerFunc: s.attachContainer},
		{Method: http.MethodGet, Path: "/containers/json", HandlerFunc: s.getContainers},
		{Method: http.MethodGet, Path: "/containers/{name:.*}/json", HandlerFunc: s.getContainer},
		{Method: http.MethodDelete, Path: "/containers/{name:.*}", HandlerFunc: s.removeContainers},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/exec", HandlerFunc: s.createContainerExec},
		{Method: http.MethodGet, Path: "/exec/{name:.*}/json", HandlerFunc: s.getExecInfo},
		{Method: http.MethodPost, Path: "/exec/{name:.*}/start", HandlerFunc: s.startContainerExec},
		{Method: http.MethodPost, Path: "/exec/{name:.*}/resize", HandlerFunc: s.resizeExec},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/rename", HandlerFunc: s.renameContainer},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/restart", HandlerFunc: s.restartContainer},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/pause", HandlerFunc: s.pauseContainer},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/unpause", HandlerFunc: s.unpauseContainer},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/update", HandlerFunc: s.updateContainer},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/upgrade", HandlerFunc: s.upgradeContainer},
		{Method: http.MethodGet, Path: "/containers/{name:.*}/top", HandlerFunc: s.topContainer},
		{Method: http.MethodGet, Path: "/containers/{name:.*}/logs", HandlerFunc: withCancelHandler(s.logsContainer)},
		{Method: http.MethodGet, Path: "/containers/{name:.*}/stats", HandlerFunc: withCancelHandler(s.statsContainer)},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/resize", HandlerFunc: s.resizeContainer},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/restart", HandlerFunc: s.restartContainer},
		{Method: http.MethodPost, Path: "/containers/{name:.*}/wait", HandlerFunc: withCancelHandler(s.waitContainer)},
		{Method: http.MethodPost, Path: "/commit", HandlerFunc: withCancelHandler(s.commitContainer)},

		// image
		{Method: http.MethodPost, Path: "/images/create", HandlerFunc: s.pullImage},
		{Method: http.MethodGet, Path: "/images/search", HandlerFunc: s.searchImages},
		{Method: http.MethodGet, Path: "/images/json", HandlerFunc: s.listImages},
		{Method: http.MethodDelete, Path: "/images/{name:.*}", HandlerFunc: s.removeImage},
		{Method: http.MethodGet, Path: "/images/{name:.*}/json", HandlerFunc: s.getImage},
		{Method: http.MethodPost, Path: "/images/{name:.*}/tag", HandlerFunc: s.postImageTag},
		{Method: http.MethodPost, Path: "/images/load", HandlerFunc: withCancelHandler(s.loadImage)},
		{Method: http.MethodGet, Path: "/images/save", HandlerFunc: withCancelHandler(s.saveImage)},
		{Method: http.MethodGet, Path: "/images/{name:.*}/history", HandlerFunc: s.getImageHistory},
		{Method: http.MethodPost, Path: "/images/{name:.*}/push", HandlerFunc: s.pushImage},

		// volume
		{Method: http.MethodGet, Path: "/volumes", HandlerFunc: s.listVolume},
		{Method: http.MethodPost, Path: "/volumes/create", HandlerFunc: s.createVolume},
		{Method: http.MethodGet, Path: "/volumes/{name:.*}", HandlerFunc: s.getVolume},
		{Method: http.MethodDelete, Path: "/volumes/{name:.*}", HandlerFunc: s.removeVolume},

		// network
		{Method: http.MethodGet, Path: "/networks", HandlerFunc: s.listNetwork},
		{Method: http.MethodPost, Path: "/networks/create", HandlerFunc: s.createNetwork},
		{Method: http.MethodGet, Path: "/networks/{id:.*}", HandlerFunc: s.getNetwork},
		{Method: http.MethodDelete, Path: "/networks/{id:.*}", HandlerFunc: s.deleteNetwork},
		{Method: http.MethodPost, Path: "/networks/{id:.*}/connect", HandlerFunc: s.connectToNetwork},
		{Method: http.MethodPost, Path: "/networks/{id:.*}/disconnect", HandlerFunc: s.disconnectNetwork},

		// metrics
		{Method: http.MethodGet, Path: "/metrics", HandlerFunc: s.metrics},

		// cri stream
		{Method: http.MethodGet, Path: "/exec/{token}", HandlerFunc: s.criExec},
		{Method: http.MethodPost, Path: "/exec/{token}", HandlerFunc: s.criExec},
		{Method: http.MethodGet, Path: "/attach/{token}", HandlerFunc: s.criAttach},
		{Method: http.MethodPost, Path: "/attach/{token}", HandlerFunc: s.criAttach},
		{Method: http.MethodGet, Path: "/portforward/{token}", HandlerFunc: s.criPortForward},
		{Method: http.MethodPost, Path: "/portforward/{token}", HandlerFunc: s.criPortForward},

		// copy
		{Method: http.MethodPut, Path: "/containers/{name:.*}/archive", HandlerFunc: s.putContainersArchive},
		{Method: http.MethodHead, Path: "/containers/{name:.*}/archive", HandlerFunc: s.headContainersArchive},
		{Method: http.MethodGet, Path: "/containers/{name:.*}/archive", HandlerFunc: s.getContainersArchive},
	}

	if s.APIPlugin != nil {
		handlers = s.APIPlugin.UpdateHandler(handlers)
	}

	// register API
	for _, h := range handlers {
		if h != nil {
			r.Path(versionMatcher + h.Path).Methods(h.Method).Handler(filter(h.HandlerFunc, s))
			r.Path(h.Path).Methods(h.Method).Handler(filter(h.HandlerFunc, s))
		}
	}

	if s.Config.Debug || s.Config.EnableProfiler {
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

// withCancelHandler will use context to cancel the handler. Otherwise, if the
// the connection has been cut by the client or firewall, the server handler
// will hang and cause goroutine leak.
func withCancelHandler(h serverTypes.Handler) serverTypes.Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		notifier, ok := rw.(http.CloseNotifier)
		if !ok {
			return h(ctx, rw, req)
		}

		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)

		waitCh := make(chan struct{})
		defer close(waitCh)

		// NOTE: in order to avoid the race , we should get the
		// channel before select.
		//
		// Related issue: https://github.com/grpc-ecosystem/grpc-gateway/pull/120.
		closeNotify := notifier.CloseNotify()
		go func() {
			select {
			case <-closeNotify:
				cancel()
			case <-waitCh:
			}
		}()
		return h(ctx, rw, req)
	}
}

func filter(handler serverTypes.Handler, s *Server) http.HandlerFunc {
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
			ctx = utils.SetTLSIssuer(ctx, issuer)
			ctx = utils.SetTLSCommonName(ctx, clientName)
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
		logrus.Errorf("Handler for %s %s, client %s returns error: %s", req.Method, req.URL.RequestURI(), clientInfo, err)
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
	} else if errtypes.IsNotModified(err) {
		code = http.StatusNotModified
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
