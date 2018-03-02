package stream

import (
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/alibaba/pouch/cri/stream/constant"
	"github.com/alibaba/pouch/cri/stream/remotecommand"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	runtimeapi "k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

// Keep these constants consistent with the peers in official package:
// k8s.io/kubernetes/pkg/kubelet/server.
const (
	// DefaultStreamIdleTimeout is the timeout for idle stream.
	DefaultStreamIdleTimeout = 4 * time.Hour

	// DefaultStreamCreationTimeout is the timeout for stream creation.
	DefaultStreamCreationTimeout = 30 * time.Second
)

// TODO: StreamProtocolV2Name, StreamProtocolV3Name, StreamProtocolV4Name support.

// SupportedStreamingProtocols is the streaming protocols which server supports.
var SupportedStreamingProtocols = []string{constant.StreamProtocolV1Name}

// SupportedPortForwardProtocols is the portforward protocols which server supports.
var SupportedPortForwardProtocols = []string{constant.PortForwardProtocolV1Name}

// Server as an interface defines all operations against stream server.
type Server interface {
	// GetExec get the serving URL for Exec request.
	GetExec(*runtimeapi.ExecRequest) (*runtimeapi.ExecResponse, error)

	// GetAttach get the serving URL for Attach request.
	GetAttach(*runtimeapi.AttachRequest) (*runtimeapi.AttachResponse, error)

	// GetPortForward get the serving URL for PortForward request.
	GetPortForward(*runtimeapi.PortForwardRequest) (*runtimeapi.PortForwardResponse, error)

	// Start starts the stream server.
	Start() error
}

// Runtime is the interface to execute the commands and provide the streams.
type Runtime interface {
	// Exec executes the command in pod.
	Exec() error

	// Attach attaches to pod.
	Attach() error

	// PortForward forward port to pod.
	PortForward() error
}

// Config defines the options used for running the stream server.
type Config struct {
	// Address is the addr:port address the server will listen on.
	Address string

	// BaseURL is the optional base URL for constructing streaming URLs. If empty, the baseURL will be constructed from the serve address.
	BaseURL *url.URL

	// StreamIdleTimeout is how long to leave idle connections open for.
	StreamIdleTimeout time.Duration
	// StreamCreationTimeout is how long to wait for clients to create streams. Only used for SPDY streaming.
	StreamCreationTimeout time.Duration

	// SupportedStreamingProtocols is the streaming protocols which server supports.
	SupportedRemoteCommandProtocols []string
	// SupportedPortForwardProtocol is the portforward protocols which server supports.
	SupportedPortForwardProtocols []string
}

// DefaultConfig provides default values for server Config.
var DefaultConfig = Config{
	StreamIdleTimeout:               4 * time.Hour,
	StreamCreationTimeout:           DefaultStreamCreationTimeout,
	SupportedRemoteCommandProtocols: SupportedStreamingProtocols,
	SupportedPortForwardProtocols:   SupportedPortForwardProtocols,
}

type server struct {
	config  Config
	runtime Runtime
	cache   *requestCache
	server  *http.Server
}

// NewServer creates a new stream server.
func NewServer(config Config, runtime Runtime) (Server, error) {
	s := &server{
		config:  config,
		runtime: runtime,
		cache:   newRequestCache(),
	}

	if s.config.BaseURL == nil {
		s.config.BaseURL = &url.URL{
			Scheme: "http",
			Host:   s.config.Address,
		}
	}

	endpoints := []struct {
		path    string
		handler http.HandlerFunc
	}{
		{"/exec/{token}", s.serveExec},
		{"/attach/{token}", s.serveAttach},
		{"/portforward{token}", s.servePortForward},
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

func (s *server) serveExec(w http.ResponseWriter, r *http.Request) {
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

func (s *server) serveAttach(w http.ResponseWriter, r *http.Request) {
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

func (s *server) servePortForward(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	cachedRequest, ok := s.cache.Consume(token)
	if !ok {
		http.NotFound(w, r)
		return
	}
	_, ok = cachedRequest.(*runtimeapi.PortForwardRequest)
	if !ok {
		http.NotFound(w, r)
		return
	}
	WriteError(grpc.Errorf(codes.NotFound, "servePortForward Has Not Been Completed Yet"), w)
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
