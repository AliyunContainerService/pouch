package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	"github.com/sirupsen/logrus"
)

// challengeClient defines a client to manage challenge–response authentication.
type challengeClient struct {
	realm   string
	service string
	scope   string
	httpCli *http.Client
}

type registryEndpoint struct {
	version string
	url     *url.URL
	// TODO: support tls for registry.
}

// token defines a token that registry may return after login successfully.
type token struct {
	Token string `json:"token"`
}

// Auth authenticates the v1/v2 registry with the credentials.
func (client *Client) Auth(config *types.AuthConfig) (token string, err error) {
	address := config.ServerAddress
	if address != "" {
		address = strings.TrimSuffix(config.ServerAddress, "/")
	}

	endpoints := genRegistryEndpoints(address)
	for _, endpoint := range endpoints {
		if endpoint.version == "v2" {
			token, err = loginV2(endpoint.url, config.Username, config.Password)
			if err == nil {
				return
			}
		} else {
			token, err = loginV1(endpoint.url, config.Username, config.Password)
			if err == nil {
				return
			}
		}
	}

	return "", err
}

// TODO: add tls support
func genRegistryEndpoints(addr string) (endpoints []registryEndpoint) {
	// add v2 registry endpoint first.
	ep := genV2Endpoints(addr)
	if ep != (registryEndpoint{}) {
		endpoints = append(endpoints, ep)
	}

	// TODO: add v1 registry endpoint

	return endpoints
}

func genV2Endpoints(addr string) registryEndpoint {
	if addr == "" {
		addr = defaultV2Registry
	}
	if !strings.HasPrefix(addr, "http") && !strings.HasPrefix(addr, "https") {
		addr = "https://" + addr
	}
	url, err := url.Parse(addr + "/v2")
	if err != nil {
		return registryEndpoint{}
	}

	return registryEndpoint{
		version: "v2",
		url:     url,
	}
}

// loginV1 login to a v1 registry with provided credential.
func loginV1(url *url.URL, username, password string) (string, error) {
	logrus.Infof("attempt to login v1 registry %s", url.String())
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return "", err
	}

	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	v1AuthClient := &http.Client{}

	resp, err := v1AuthClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to login v1 registry %s with http status %s", url.String(), http.StatusText(resp.StatusCode))
	}
	token := token{}
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return "", err
	}

	return token.Token, err
}

// pingV2 wants to get challenge from http response, since
// v2 registry uses challenge–response authentication.
func pingV2(url *url.URL) ([]Challenge, error) {
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	challenges := parseAuthHeader(resp.Header)
	if len(challenges) == 0 {
		return nil, fmt.Errorf("failed to get challenge message in ping v2 registry")
	}

	return challenges, nil
}

func newV2AuthClient(challenges []Challenge) *challengeClient {
	client := &challengeClient{
		httpCli: &http.Client{},
	}

	for _, c := range challenges {
		if c.Scheme == "bearer" {
			if realm, ex1 := c.Parameters["realm"]; ex1 {
				client.realm = realm
			}
			if service, ex2 := c.Parameters["service"]; ex2 {
				client.service = service
			}
			if scope, ex3 := c.Parameters["scope"]; ex3 {
				client.scope = scope
			}
			break
		}
	}

	return client
}

// loginV2 login to a v2 registry with provided credential.
func loginV2(url *url.URL, username, password string) (string, error) {
	logrus.Infof("attempt to login v2 registry %s", url.String())
	challenges, err := pingV2(url)
	if err != nil {
		return "", err
	}

	v2AuthClient := newV2AuthClient(challenges)

	realmURL, err := url.Parse(v2AuthClient.realm)
	if err != nil {
		return "", err
	}

	q := url.Query()
	if v2AuthClient.service != "" {
		q.Set("service", v2AuthClient.service)
	}
	if v2AuthClient.scope != "" {
		q.Set("scope", v2AuthClient.scope)
	}
	realmURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", realmURL.String(), nil)
	if err != nil {
		return "", err
	}

	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := v2AuthClient.httpCli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to login v2 registry %s with http status %s", url.String(), http.StatusText(resp.StatusCode))
	}
	token := token{}
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return "", err
	}

	return token.Token, err
}
