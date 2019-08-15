package plugins

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/alibaba/pouch/pkg/httputils"
	"github.com/alibaba/pouch/pkg/log"
)

var (
	maxRequestTimeout  = 30 * time.Second
	defaultDialTimeout = 10 * time.Second
	defaultCliTimeout  = 600 * time.Second
)

// defaultContentType is the default Content-Type accepted and sent by the plugins.
const defaultContentType = "application/vnd.docker.plugins.v1.1+json"

// PluginClient is the plugin client.
type PluginClient struct {
	// address is the plugin server address
	address string
	// baseURL is the request base url
	baseURL string
	// the http client
	client *http.Client
}

// NewPluginClient creates a PluginClient from the address and tls config.
func NewPluginClient(addr string, tlsconfig *TLSConfig) (*PluginClient, error) {
	var config *tls.Config
	var baseURL string
	var err error

	if addr == "" {
		return nil, fmt.Errorf("empty plugin address is invalid")
	}

	url, baseURL, _, err := httputils.ParseHost(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plugin address %s: %v", addr, err)
	}

	// Scheme is https, generate the tls config
	if url.Scheme == "https" {
		if tlsconfig != nil && !tlsconfig.InsecureSkipVerify {
			config, err = httputils.GenTLSConfig(tlsconfig.KeyFile, tlsconfig.CertFile, tlsconfig.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to parse plugin tls config: %v", err)
			}
		} else {
			config = &tls.Config{InsecureSkipVerify: true}
		}
	}

	httpCli := httputils.NewHTTPClient(url, config, defaultDialTimeout, defaultCliTimeout)

	return &PluginClient{
		address: addr,
		baseURL: baseURL,
		client:  httpCli,
	}, nil
}

// newPluginRequest generates a plugin request
func (cli *PluginClient) newPluginRequest(path string, data io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	fullPath := cli.baseURL + path

	request, err := http.NewRequest(http.MethodPost, fullPath, data)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", defaultContentType)

	return request, nil
}

// CallService calls the service provided by plugin server.
func (cli *PluginClient) CallService(service string, in, out interface{}, retry bool) error {
	input := new(bytes.Buffer)
	if in != nil {
		if err := json.NewEncoder(input).Encode(in); err != nil {
			return err
		}
	}

	resp, err := cli.callService(service, input, retry)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &ErrPluginStatus{
				StatusCode: resp.StatusCode,
				Message:    resp.Status,
			}
		}

		return &ErrPluginStatus{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	if out == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (cli *PluginClient) callService(service string, data io.Reader, retry bool) (*http.Response, error) {
	var start = time.Now()
	var times = 0

	// generate the request.
	req, err := cli.newPluginRequest(service, data)
	if err != nil {
		return nil, err
	}

	for {
		resp, err := cli.client.Do(req)
		if err != nil {
			if !retry {
				return nil, err
			}
			delay := backoff(times)
			if timeout(start, delay) {
				return nil, err
			}
			times++
			log.With(nil).Warnf("plugin %s call %s failed, retry after %v seconds", cli.address, service, delay.Seconds())
			time.Sleep(delay)
			continue
		}

		return resp, nil
	}
}

// backoff returns the delay time.
func backoff(times int) time.Duration {
	b := time.Second

	for times > 0 && b < maxRequestTimeout {
		b *= 2
		times--
	}

	if b > maxRequestTimeout {
		b = maxRequestTimeout
	}

	return b
}

// timeout checks whether timeout.
func timeout(start time.Time, delay time.Duration) bool {
	return delay+time.Since(start) >= maxRequestTimeout
}
