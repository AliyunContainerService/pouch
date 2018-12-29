package v1alpha2

import (
	"net/http"
	"net/url"
	"path"

	runtimeapi "github.com/alibaba/pouch/cri/apis/v1alpha2"
	"github.com/alibaba/pouch/cri/stream"
	"github.com/alibaba/pouch/cri/stream/portforward"
	"github.com/alibaba/pouch/cri/stream/remotecommand"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// StreamServer as an interface defines all operations against stream server.
type StreamServer interface {
	// GetExec get the serving URL for Exec request.
	GetExec(*runtimeapi.ExecRequest) (*runtimeapi.ExecResponse, error)

	// GetAttach get the serving URL for Attach request.
	GetAttach(*runtimeapi.AttachRequest) (*runtimeapi.AttachResponse, error)

	// GetPortForward get the serving URL for PortForward request.
	GetPortForward(*runtimeapi.PortForwardRequest) (*runtimeapi.PortForwardResponse, error)

	// Start starts the stream server.
	Start() error

	// Router is the Stream Server's handlers which we should export.
	stream.Router
}

type server struct {
	config  stream.Config
	runtime stream.Runtime
	cache   *stream.RequestCache
	server  *http.Server
}

// NewStreamServer creates a new stream server.
func NewStreamServer(config stream.Config, runtime stream.Runtime) (StreamServer, error) {
	s := &server{
		config:  config,
		runtime: runtime,
		cache:   stream.NewRequestCache(),
	}

	endpoints := []struct {
		path    string
		handler http.HandlerFunc
	}{
		{"/exec/{token}", s.ServeExec},
		{"/attach/{token}", s.ServeAttach},
		{"/portforward/{token}", s.ServePortForward},
	}

	r := mux.NewRouter()
	for _, e := range endpoints {
		for _, method := range []string{"GET", "POST"} {
			r.Path(e.path).Methods(method).Handler(e.handler)
		}
	}

	s.server = &http.Server{
		Addr:    s.config.Address,
		Handler: r,
	}

	return s, nil
}

// Start starts the stream server.
func (s *server) Start() error {
	return s.server.ListenAndServe()
}

func (s *server) ServeExec(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := mux.Vars(r)["token"]
	cachedRequest, ok := s.cache.Consume(token)
	if !ok {
		http.NotFound(w, r)
		return
	}

	exec, ok := cachedRequest.(*runtimeapi.ExecRequest)
	if !ok {
		http.NotFound(w, r)
		return
	}

	streamOpts := &remotecommand.Options{
		Stdin:  exec.Stdin,
		Stdout: exec.Stdout,
		Stderr: exec.Stderr,
		TTY:    exec.Tty,
	}
	remotecommand.ServeExec(
		ctx,
		w,
		r,
		s.runtime,
		exec.ContainerId,
		exec.Cmd,
		streamOpts,
		s.config.SupportedRemoteCommandProtocols,
		s.config.StreamIdleTimeout,
		s.config.StreamCreationTimeout,
	)
}

func (s *server) ServeAttach(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := mux.Vars(r)["token"]
	cachedRequest, ok := s.cache.Consume(token)
	if !ok {
		http.NotFound(w, r)
		return
	}
	attach, ok := cachedRequest.(*runtimeapi.AttachRequest)
	if !ok {
		http.NotFound(w, r)
		return
	}

	streamOpts := &remotecommand.Options{
		Stdin:  attach.Stdin,
		Stdout: attach.Stdout,
		Stderr: attach.Stderr,
		TTY:    attach.Tty,
	}
	remotecommand.ServeAttach(
		ctx,
		w,
		r,
		s.runtime,
		attach.ContainerId,
		streamOpts,
		s.config.StreamIdleTimeout,
		s.config.StreamCreationTimeout,
		s.config.SupportedRemoteCommandProtocols,
	)
}

func (s *server) ServePortForward(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := mux.Vars(r)["token"]
	cachedRequest, ok := s.cache.Consume(token)
	if !ok {
		http.NotFound(w, r)
		return
	}
	pf, ok := cachedRequest.(*runtimeapi.PortForwardRequest)
	if !ok {
		http.NotFound(w, r)
		return
	}

	portforward.ServePortForward(
		ctx,
		w,
		r,
		s.runtime,
		pf.PodSandboxId,
		s.config.StreamIdleTimeout,
		s.config.StreamCreationTimeout,
		s.config.SupportedPortForwardProtocols,
	)
}

func (s *server) buildURL(method string, token string) string {
	return s.config.BaseURL.ResolveReference(&url.URL{
		Path: path.Join(method, token),
	}).String()
}

// GetExec gets the serving URL for the Exec requests.
func (s *server) GetExec(req *runtimeapi.ExecRequest) (*runtimeapi.ExecResponse, error) {
	// TODO: validate the request.
	token, err := s.cache.Insert(req)
	if err != nil {
		return nil, err
	}

	return &runtimeapi.ExecResponse{
		Url: s.buildURL("exec", token),
	}, nil
}

// GetAttach gets the serving URL for the Attach requests.
func (s *server) GetAttach(req *runtimeapi.AttachRequest) (*runtimeapi.AttachResponse, error) {
	// TODO: validate the request.
	token, err := s.cache.Insert(req)
	if err != nil {
		return nil, err
	}

	return &runtimeapi.AttachResponse{
		Url: s.buildURL("attach", token),
	}, nil
}

// GetPortForward gets the serving URL for the PortForward requests.
func (s *server) GetPortForward(req *runtimeapi.PortForwardRequest) (*runtimeapi.PortForwardResponse, error) {
	if req.PodSandboxId == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "missing required pod_sandbox_id")
	}
	token, err := s.cache.Insert(req)
	if err != nil {
		return nil, err
	}

	return &runtimeapi.PortForwardResponse{
		Url: s.buildURL("portforward", token),
	}, nil
}
