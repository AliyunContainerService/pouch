package request

import (
	"io"
	"net/http"

	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/test/environment"
)

// Get sends request to the default pouchd server with custom request options.
func Get(endpoint string, opts ...func(*http.Request)) (*http.Response, error) {
	apiClient, err := newAPIClient(environment.PouchdAddress, environment.TLSConfig)
	if err != nil {
		return nil, err
	}

	req, err := newRequest(http.MethodGet, apiClient.BaseURL()+endpoint, nil, opts...)
	if err != nil {
		return nil, err
	}
	return apiClient.HTTPCli.Do(req)
}

// newAPIClient return new HTTP client with tls.
//
// FIXME: Could we make some functions exported in alibaba/pouch/client?
func newAPIClient(host string, tls utils.TLSConfig) (*client.APIClient, error) {
	return client.NewAPIClient(host, tls)
}

// newRequest creates request targeting on specific host/path by method.
func newRequest(method, url string, body io.Reader, opts ...func(*http.Request)) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}
	return req, nil
}
