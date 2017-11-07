package client

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testHost = "unix:///var/run/pouchd.sock"
)

func newClient(t *testing.T) *Client {
	client, err := New("")
	if err != nil {
		t.Fatal("failed to new client: ", err)
	}

	return client
}

func TestNewClient(t *testing.T) {
	assert := assert.New(t)
	kvs := map[string]bool{
		"":                      false,
		"foobar":                true,
		"tcp://localhost:2476":  false,
		"http://localhost:2476": false,
	}

	for host, expectError := range kvs {
		cli, err := New(host)
		if expectError {
			assert.Error(err, fmt.Sprintf("test data %v", host))
		} else {
			assert.NoError(err, fmt.Sprintf("test data %v", host))
		}

		t.Logf("client info %+v", cli)
	}
}

func TestParseHost(t *testing.T) {
	assert := assert.New(t)
	type parsed struct {
		host           string
		expectError    bool
		expectBasePath string
	}

	parseds := []parsed{
		{host: testHost, expectError: false, expectBasePath: "http://d"},
		{host: "tcp://localhost:1234", expectError: false, expectBasePath: "http://localhost:1234"},
		{host: "http://localhost:5678", expectError: false, expectBasePath: "http://localhost:5678"},
		{host: "foo:bar", expectError: true, expectBasePath: ""},
		{host: "", expectError: true, expectBasePath: ""},
	}

	for _, p := range parseds {
		_, basePath, err := parseHost(p.host)
		if p.expectError {
			assert.Error(err, fmt.Sprintf("test data %v", p.host))
		} else {
			assert.NoError(err, fmt.Sprintf("test data %v", p.host))
		}

		assert.Equal(basePath, p.expectBasePath)
	}
}
