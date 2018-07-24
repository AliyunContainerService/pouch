package client

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testHost = "unix:///var/run/pouchd.sock"
)

func TestNewAPIClient(t *testing.T) {
	assert := assert.New(t)
	kvs := map[string]bool{
		"":                      false,
		"foobar":                true,
		"tcp://localhost:2476":  false,
		"http://localhost:2476": false,
	}

	for host, expectError := range kvs {
		cli, err := NewAPIClient(host, TLSConfig{})
		if expectError {
			assert.Error(err, fmt.Sprintf("test data: %v", host))
		} else {
			assert.NoError(err, fmt.Sprintf("test data %v: %v", host, err))
		}

		t.Logf("client info %+v", cli)
	}
}

func Test_generateBaseURL(t *testing.T) {
	type args struct {
		u   *url.URL
		tls TLSConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateBaseURL(tt.args.u, tt.args.tls); got != tt.want {
				t.Errorf("generateBaseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
