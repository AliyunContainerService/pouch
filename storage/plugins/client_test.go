package plugins

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
)

func setupPluginServer() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
}

func teardownPluginServer() {
	if server != nil {
		server.Close()
	}
}

func TestNewPluginClient(t *testing.T) {
	cases := []struct {
		address string
		success bool
		baseURL string
	}{
		{"unix:///var/run/plugins/vol1.sock", true, "http://d"},
		{"tcp://127.0.0.1:8001", true, "http://127.0.0.1:8001"},
		{"http://127.0.0.1:8002", true, "http://127.0.0.1:8002"},
		{"https://127.0.0.1:8003", true, "https://127.0.0.1:8003"},
		{"127.0.0.1:8004", false, ""},
	}

	for _, c := range cases {
		client, err := NewPluginClient(c.address, nil)

		// expect failed
		if !c.success && err == nil {
			t.Fatalf("NewPluginClient from %s success, but expect failed", c.address)
		}

		if c.success {
			if err != nil {
				t.Fatalf("NewPluginClient from %s failed, but expect success", c.address)
			}
			if c.baseURL != client.baseURL {
				t.Fatalf("NewPluginClient get baseURL: %s, but expect %s", client.baseURL, c.baseURL)
			}
		}
	}
}

func TestBackoff(t *testing.T) {
	cases := []struct {
		counter int
		delay   time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
		{5, 30 * time.Second},
		{6, 30 * time.Second},
	}

	for _, c := range cases {
		d := backoff(c.counter)
		if d != c.delay {
			t.Fatalf("backoff expected %v delay, but got %v", c.delay, d)
		}
	}
}

func TestCallService(t *testing.T) {
	setupPluginServer()
	defer teardownPluginServer()

	address := server.URL

	input := HandShakeResp{
		Implements: []string{"VolumeDriver"},
	}

	mux.HandleFunc("/testService", func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		contentType := r.Header.Get("Content-Type")

		if method != http.MethodPost {
			t.Fatalf("PluginServer expect %s method, but got %s method",
				http.MethodPost, r.Method)
		}

		if contentType != defaultContentType {
			t.Fatalf("PluginServer expect Content-Type is %s,but got %s",
				defaultContentType, contentType)
		}

		io.Copy(w, r.Body)

	})

	cli, err := NewPluginClient(address, nil)
	if err != nil {
		t.Fatal(err)
	}

	var output HandShakeResp
	err = cli.CallService("/testService", &input, &output, true)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(input, output) {
		t.Fatalf("expect %v, but got %v", input, output)
	}
}
