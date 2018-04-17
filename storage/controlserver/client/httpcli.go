package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"

	"github.com/alibaba/pouch/pkg/serializer"

	"github.com/go-resty/resty"
)

// HTTPClient represents a client connect to http server.
type HTTPClient struct {
	cli    *resty.Client
	req    *resty.Request
	err    error
	resp   *resty.Response
	method string
	urls   string
	code   int
}

var restyClient = resty.New().SetCloseConnection(true)

func isPtr(i interface{}) bool {
	return reflect.ValueOf(i).Kind() == reflect.Ptr
}

func parseTLSConfig(ca, cert, key []byte) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		// Prefer TLS1.2 as the client minimum
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: os.Getenv("DOCKER_TLS_VERIFY") == "",
	}

	//parse root ca
	cas, err := certpool(ca)
	if err != nil {
		return nil, err
	}

	tlsConfig.RootCAs = cas

	//parse cert and key pem
	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	tlsConfig.Certificates = []tls.Certificate{tlsCert}
	return tlsConfig, nil
}

func certpool(pem []byte) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("failed to append certificates from PEM")
	}

	return certPool, nil
}

// HTTPClientNew returns HTTPClient class object.
func HTTPClientNew() *HTTPClient {
	return &HTTPClient{
		cli:  restyClient,
		req:  restyClient.R(),
		resp: &resty.Response{},
	}
}

func (c *HTTPClient) verb(method string) *HTTPClient {
	c.method = method
	return c
}

// TLSConfig is used to set tls config with specified config.
func (c *HTTPClient) TLSConfig(tlsc *tls.Config) *HTTPClient {
	c.cli.SetTLSClientConfig(tlsc)
	return c
}

// TLS is used to set tls config with ca, cert and key.
func (c *HTTPClient) TLS(ca, cert, key []byte) (*HTTPClient, error) {
	tlsc, err := parseTLSConfig(ca, cert, key)
	if err != nil {
		return nil, err
	}
	c.cli.SetTLSClientConfig(tlsc)
	return c, nil
}

// PUT used to implement http PUT method.
func PUT() *HTTPClient {
	c := HTTPClientNew()
	return c.verb("PUT")
}

// GET used to implement HTTPClient GET method.
func GET() *HTTPClient {
	c := HTTPClientNew()
	return c.verb("GET")
}

// POST used to implement HTTPClient POST method.
func POST() *HTTPClient {
	c := HTTPClientNew()
	return c.verb("POST")
}

// DELETE used to implement HTTPClient DELETE method.
func DELETE() *HTTPClient {
	c := HTTPClientNew()
	return c.verb("DELETE")
}

// PUT used to implement HTTPClient PUT method and then returns HTTPClient
func (c *HTTPClient) PUT() *HTTPClient {
	c.verb("PUT")
	return c
}

// GET used to implement HTTPClient GET method and then returns HTTPClient
func (c *HTTPClient) GET() *HTTPClient {
	c.verb("GET")
	return c
}

// POST used to implement HTTPClient POST method and then returns HTTPClient
func (c *HTTPClient) POST() *HTTPClient {
	c.verb("POST")
	return c
}

// DELETE used to implement HTTPClient DELETE method and then returns HTTPClient
func (c *HTTPClient) DELETE() *HTTPClient {
	c.verb("DELETE")
	return c
}

// Err returns HTTPClient's error.
func (c *HTTPClient) Err() error {
	return c.err
}

// Method used to implement HTTPClient method that specified by caller,
// and then returns HTTPClient object.
func (c *HTTPClient) Method(method string) *HTTPClient {
	c.verb(method)
	return c
}

// URL is used to set url and then returns HTTPClient object.
func (c *HTTPClient) URL(rawurl string) *HTTPClient {
	if _, err := url.Parse(rawurl); err != nil {
		c.err = err
		return c
	}
	c.urls = rawurl
	return c
}

// SetHeader is used to set header and then returns HTTPClient object.
func (c *HTTPClient) SetHeader(key, value string) *HTTPClient {
	c.req.SetHeader(key, value)
	return c
}

// Body is used to set body with map[string]string type,
// and then returns HTTPClient object.
func (c *HTTPClient) Body(obj serializer.Object) *HTTPClient {
	if c.err != nil {
		return c
	}
	switch t := obj.(type) {
	case map[string]string:
		c.req.SetFormData(t)
	default:
		c.err = fmt.Errorf("post body need a map[string]string, get: %v", t)
	}
	return c
}

// JSONBody is used to set body with json format and then returns HTTPClient object.
func (c *HTTPClient) JSONBody(obj serializer.Object) *HTTPClient {
	c.req.SetHeader("Content-Type", "application/json")
	c.req.SetBody(obj)
	return c
}

// Query is used to set query and then returns HTTPClient object.
func (c *HTTPClient) Query(obj serializer.Object) *HTTPClient {
	if c.err != nil {
		return c
	}
	switch t := obj.(type) {
	case string:
		c.req.SetQueryString(t)
	case map[string][]string:
		query := url.Values(t)
		c.req.SetQueryString(query.Encode())
	case map[string]string:
		c.req.SetQueryParams(t)
	case url.Values:
		c.req.SetQueryString(t.Encode())
	default:
		c.err = fmt.Errorf("invalid query %#v", obj)
	}

	return c
}

// Do used to deal http request and then returns HTTPClient object.
func (c *HTTPClient) Do() *HTTPClient {
	if c.err != nil {
		return c
	}

	switch c.method {
	case "POST":
		c.resp, c.err = c.req.Post(c.urls)
	case "GET":
		c.resp, c.err = c.req.Get(c.urls)
	case "DELETE":
		c.resp, c.err = c.req.Delete(c.urls)
	case "PUT":
		c.resp, c.err = c.req.Put(c.urls)
	default:
		c.err = fmt.Errorf("unsupport method: %s", c.method)
		return c
	}

	c.code = c.resp.StatusCode()
	if c.code >= http.StatusBadRequest {
		c.err = fmt.Errorf("Bad response code: %d, %s", c.code, string(c.resp.Body()))
	}
	return c
}

// RespCodeEqual check HTTPClient's code is equal with "code" or not.
func (c *HTTPClient) RespCodeEqual(code int) bool {
	return c.code == code
}

// StatusCode returns HTTPClient code.
func (c *HTTPClient) StatusCode() int {
	return c.code
}

// Into used to set obj into body.
func (c *HTTPClient) Into(obj serializer.Object) error {
	if c.err != nil {
		return c.err
	}
	if !isPtr(obj) {
		return fmt.Errorf("expect ptr, get: %v", obj)
	}
	body := c.resp.Body()
	if reflect.TypeOf(obj).Elem().Kind() == reflect.String {
		reflect.Indirect(reflect.ValueOf(obj)).Set(reflect.Indirect(reflect.ValueOf(string(body))))
		return nil
	}

	newObj := reflect.New(reflect.TypeOf(obj).Elem()).Interface()

	if err := serializer.Codec.Decode(body, newObj); err != nil {
		return err
	}
	reflect.Indirect(reflect.ValueOf(obj)).Set(reflect.Indirect(reflect.ValueOf(newObj)))
	return nil
}
