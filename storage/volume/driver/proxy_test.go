package driver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alibaba/pouch/storage/plugins"
)

func TestRemoteDriverRequestError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc(remoteVolumeCreateService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Err": "Cannot create volume"}`)
	})

	mux.HandleFunc(remoteVolumeRemoveService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Err": "Cannot remove volume"}`)
	})

	mux.HandleFunc(remoteVolumeMountService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Err": "Cannot mount volume"}`)
	})

	mux.HandleFunc(remoteVolumeUnmountService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Err": "Cannot unmount volume"}`)
	})

	mux.HandleFunc(remoteVolumePathService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Err": "Unknown volume"}`)
	})

	mux.HandleFunc(remoteVolumeListService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Err": "Cannot list volumes"}`)
	})

	mux.HandleFunc(remoteVolumeGetService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Err": "Cannot get volume"}`)
	})

	mux.HandleFunc(remoteVolumeCapabilitiesService, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1+json")
		http.Error(w, "error", 500)
	})

	u, _ := url.Parse(server.URL)
	client, err := plugins.NewPluginClient("tcp://"+u.Host, &plugins.TLSConfig{InsecureSkipVerify: true})
	if err != nil {
		t.Fatal(err)
	}

	dp := &remoteDriverProxy{
		client: client,
	}

	if err = dp.Create("volume", nil); err == nil {
		t.Fatal("Expected error, was nil")
	}

	if !strings.Contains(err.Error(), "Cannot create volume") {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	_, err = dp.Mount("volume", "abc")
	if err == nil {
		t.Fatal("Expected error, was nil")
	}

	if !strings.Contains(err.Error(), "Cannot mount volume") {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	err = dp.Unmount("volume", "abc")
	if err == nil {
		t.Fatal("Expected error, was nil")
	}

	if !strings.Contains(err.Error(), "Cannot unmount volume") {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	err = dp.Remove("volume")
	if err == nil {
		t.Fatal("Expected error, was nil")
	}

	if !strings.Contains(err.Error(), "Cannot remove volume") {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	_, err = dp.Path("volume")
	if err == nil {
		t.Fatal("Expected error, was nil")
	}

	if !strings.Contains(err.Error(), "Unknown volume") {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	_, err = dp.List()
	if err == nil {
		t.Fatal("Expected error, was nil")
	}
	if !strings.Contains(err.Error(), "Cannot list volumes") {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	_, err = dp.Get("volume")
	if err == nil {
		t.Fatal("Expected error, was nil")
	}
	if !strings.Contains(err.Error(), "Cannot get volume") {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	_, err = dp.Capabilities()
	if err == nil {
		t.Fatal(err)
	}
}
