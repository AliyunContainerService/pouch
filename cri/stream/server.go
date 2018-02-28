package stream

import (
	"time"
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/mux"
	runtimeapi "k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"github.com/sirupsen/logrus"
)

const (
	// Keep these constants consistent with the peers in official package:
	// k8s.io/kubernetes/pkg/kubelet/server.
	DefaultStreamIdleTimeout = 4 * time.Hour

	DefaultStreamCreationTimeout = 30 * time.Second

	// The SPDY subprotocol "channel.k8s.io" is used for remote command
	// attachment/execution. This represents the initial unversioned subprotocol,
	// which has the known bugs http://issues.k8s.io/13394 and
	// http://issues.k8s.io/13395.
	StreamProtocolV1Name = "channel.k8s.io"

	// The SPDY subprotocol "v2.channel.k8s.io" is used for remote command
	// attachment/execution. It is the second version of the subprotocol and
	// resolves the issues present in the first version.
	StreamProtocolV2Name = "v2.channel.k8s.io"

	// The SPDY subprotocol "v3.channel.k8s.io" is used for remote command
	// attachment/execution. It is the third version of the subprotocol and
	// adds support for resizing container terminals.
	StreamProtocolV3Name = "v3.channel.k8s.io"

	// The SPDY subprotocol "v4.channel.k8s.io" is used for remote command
	// attachment/execution. It is the 4th version of the subprotocol and
	// adds support for exit codes.
	StreamProtocolV4Name = "v4.channel.k8s.io"

	// The SPDY subprotocol "portforward.k8s.io" is used for port forwarding.
	PortForwardProtocolV1Name = "portforward.k8s.io"
)

var SupportedStreamingProtocols = []string{StreamProtocolV4Name, StreamProtocolV3Name, StreamProtocolV2Name, StreamProtocolV1Name}

var SupportedPortForwardProtocols = []string{PortForwardProtocolV1Name}

type Server interface {
	// Get the serving URL for the requests.
	// Requests must not be nil. Responses may be nil if an error is returned.
	GetExec(*runtimeapi.ExecRequest) (*runtimeapi.ExecResponse, error)
	GetAttach(*runtimeapi.AttachRequest) (*runtimeapi.AttachResponse, error)
	GetPortForward(*runtimeapi.PortForwardRequest) (*runtimeapi.PortForwardResponse, error)

	Start() error
}

// The interface to execute the commands and provide the streams.
type Runtime interface {
	Exec() error
	Attach() error
	PortForward() error
}

// Config defines the options used for running the stream server.
type Config struct {
	// The addr:port address the server will listen on.
	Address string
	// The optional base URL for constructing streaming URLs. If empty, the baseURL will be
	// constructed from the serve address.
	BaseURL *url.URL

	// How long to leave idle connections open for.
	StreamIdleTimeout time.Duration
	// How long to wait for clients to create streams. Only used for SPDY streaming.
	StreamCreationTimeout time.Duration

	// All the stream protocols that server supports.
	SupportedRemoteCommandProtocols []string
	SupportedPortForwardProtocols []string
}

var DefaultConfig = Config{
	StreamIdleTimeout:					4 * time.Hour,
	StreamCreationTimeout:				DefaultStreamCreationTimeout,
	SupportedRemoteCommandProtocols:	SupportedStreamingProtocols,
	SupportedPortForwardProtocols:		SupportedPortForwardProtocols,
}

type server struct {
	config 	Config
	runtime Runtime
	cache	*requestCache
	server	*http.Server	
}

func NewServer(config Config, runtime Runtime) (Server, error) {
	s := &server{
		config:		config,
		runtime:	runtime,
		cache:		newRequestCache(),
	}

	if s.config.BaseURL == nil {
		s.config.BaseURL = &url.URL{
			Scheme: "http",
			Host:   s.config.Address,
		}
	}

	endpoints := []struct {
		path	string
		handler	http.HandlerFunc
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
		Addr:		s.config.Address,
		Handler:	r,	
	}

	return s, nil
}

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

	_, ok = cachedRequest.(*runtimeapi.ExecRequest)
	if !ok {
		http.NotFound(w, r)
		return
	}

	logrus.Infof("serverExec Has Not Been Completed Not")

	WriteError(grpc.Errorf(codes.NotFound, "serveExec Has Not Been Completed Yet"), w)
}

func (s *server) serveAttach(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	cachedRequest, ok := s.cache.Consume(token)
	if !ok {
		http.NotFound(w, r)
		return
	}
	_, ok = cachedRequest.(*runtimeapi.AttachRequest)
	if !ok {
		http.NotFound(w, r)
		return
	}
	WriteError(grpc.Errorf(codes.NotFound, "serveAttach Has Not Been Completed Yet"), w)
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
		Path:	path.Join(method, token),
	}).String()
}

// Get the serving URL for the Exec requests.
func (s *server) GetExec(req *runtimeapi.ExecRequest) (*runtimeapi.ExecResponse, error) {
	// TODO: validate the request.
	token, err := s.cache.Insert(req)
	if err != nil {
		return nil, err
	}

	return &runtimeapi.ExecResponse{
		Url:	s.buildURL("exec", token),
	}, nil
}

// Get the serving URL for the Attach requests.
func (s *server) GetAttach(req *runtimeapi.AttachRequest) (*runtimeapi.AttachResponse, error) {
	// TODO: validate the request.
	token, err := s.cache.Insert(req)
	if err != nil {
		return nil, err
	}

	return &runtimeapi.AttachResponse{
		Url:	s.buildURL("attach", token),
	}, nil
}

// Get the serving URL for the PortForward requests.
func (s *server) GetPortForward(req *runtimeapi.PortForwardRequest) (*runtimeapi.PortForwardResponse, error){
	if req.PodSandboxId == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "missing required pod_sandbox_id")
	}
	token, err := s.cache.Insert(req)
	if err != nil {
		return nil, err
	}

	return &runtimeapi.PortForwardResponse{
		Url:	s.buildURL("portforward", token),
	}, nil
}
