package client

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/alibaba/pouch/pkg/serializer"
)

var (
	// VolumePath is volume path url prefix
	VolumePath = "/volume"

	// StoragePath is storage path url prefix
	StoragePath = "/storage"
)

// Client represents a client connect to control server.
type Client struct {
	ttl     time.Duration
	errChan chan error
	cli     *HTTPClient
	tlsc    *tls.Config
}

// New is used to initialize client class object.
func New() *Client {
	return &Client{
		ttl:     time.Second * 90,
		cli:     HTTPClientNew(),
		errChan: make(chan error, 1),
	}
}

// NewWithTLS is used to initialize client class object with tls.
func NewWithTLS(tlsc *tls.Config) *Client {
	c := New()
	c.cli.TLSConfig(tlsc)
	return c
}

func (c *Client) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.TODO(), c.ttl)
}

func (c *Client) wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c.errChan:
		if err != nil {
			if code := c.cli.StatusCode(); code >= http.StatusBadRequest {
				return newError(code, err.Error())
			}
		}
		return err
	}
}

// Create is used to create "obj" with url on control server.
func (c *Client) Create(url string, obj serializer.Object) error {
	ctx, cancel := c.context()
	defer cancel()
	go func(obj serializer.Object) {
		c.errChan <- c.cli.POST().URL(url).JSONBody(obj).Do().Into(obj)
	}(obj)

	return c.wait(ctx)
}

// Update is used to update "obj" with url on control server.
func (c *Client) Update(url string, obj serializer.Object) error {
	ctx, cancel := c.context()
	defer cancel()
	go func(obj serializer.Object) {
		c.errChan <- c.cli.PUT().URL(url).JSONBody(obj).Do().Into(obj)
	}(obj)

	return c.wait(ctx)
}

// Get returns "obj" content with url on control server.
func (c *Client) Get(url string, obj serializer.Object) error {
	ctx, cancel := c.context()
	defer cancel()
	go func(obj serializer.Object) {
		c.errChan <- c.cli.GET().URL(url).Do().Into(obj)
	}(obj)

	return c.wait(ctx)
}

// Delete is used to delete "obj" with url on control server.
func (c *Client) Delete(url string, obj serializer.Object) error {
	ctx, cancel := c.context()
	defer cancel()
	go func() {
		c.errChan <- c.cli.DELETE().URL(url).JSONBody(obj).Do().Err()
	}()

	return c.wait(ctx)
}

// ListKeys returns "obj" with url that contains labels or keys.
func (c *Client) ListKeys(url string, obj serializer.Object) error {
	ctx, cancel := c.context()
	defer cancel()
	go func(obj serializer.Object) {
		c.errChan <- c.cli.GET().URL(url).Do().Into(obj)
	}(obj)

	return c.wait(ctx)
}
