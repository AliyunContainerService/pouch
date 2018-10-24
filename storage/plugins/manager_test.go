package plugins

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/alibaba/pouch/pkg/utils"
)

func setupPluginDir() (string, error) {
	tmpDir, err := ioutil.TempDir("", "pouch_plugins")
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return "", err
	}

	return tmpDir, nil
}

func TestNewPluginFromAddr(t *testing.T) {
	tmpDir, err := setupPluginDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := &pluginManager{
		plugins:         make(map[string]*Plugin),
		pluginSockPaths: []string{tmpDir},
		pluginSpecPaths: []string{tmpDir},
	}

	pName := "example"
	specPath := path.Join(tmpDir, fmt.Sprintf("%s.spec", pName))
	specContent := "unix:///var/run/pouch/plugins/example.sock"

	err = ioutil.WriteFile(specPath, []byte(specContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	pl, err := m.newPlugin(pName)
	if err != nil {
		t.Fatal(err)
	}

	if pl.Name != pName {
		t.Fatalf("NewPluginFromAddr expect to get %s plugin, but got %s", pName, pl.Name)
	}
}

func TestNewPluginFromJSON(t *testing.T) {
	tmpDir, err := setupPluginDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := &pluginManager{
		plugins:         make(map[string]*Plugin),
		pluginSockPaths: []string{tmpDir},
		pluginSpecPaths: []string{tmpDir},
	}

	pName := "example"
	specPath := path.Join(tmpDir, fmt.Sprintf("%s.json", pName))
	specContent := `
    {
        "Name": "example",
        "Addr": "https://example.com/pouch/plugin"
    }
    `

	err = ioutil.WriteFile(specPath, []byte(specContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	pl, err := m.newPlugin(pName)
	if err != nil {
		t.Fatal(err)
	}

	if pl.Name != pName {
		t.Fatalf("NewPluginFromJSON expect to get %s plugin, but got %s", pName, pl.Name)
	}
}

func TestGetPluginByName(t *testing.T) {
	setupPluginServer()
	defer teardownPluginServer()

	tmpDir, err := setupPluginDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := &pluginManager{
		plugins:         make(map[string]*Plugin),
		pluginSockPaths: []string{tmpDir},
		pluginSpecPaths: []string{tmpDir},
	}

	// plugin name.
	pName := "example"
	// the example plugin implements volume.
	implements := []string{"volume"}
	specPath := path.Join(tmpDir, fmt.Sprintf("%s.spec", pName))

	// write the spec.
	err = ioutil.WriteFile(specPath, []byte(server.URL), 0644)
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(HandShakePath, func(w http.ResponseWriter, r *http.Request) {
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

		// the plugin implement volume.
		resp := HandShakeResp{implements}

		content, err := json.Marshal(resp)
		if err != nil {
			t.Fatal(err)
		}

		w.Write(content)
	})

	// get the "example" plugin, expect it implements volume.
	_, err = m.getPluginByName("volume", pName)
	if err != nil {
		t.Fatalf("expect get example plugin, which implements volume, but got: %v", err)
	}

	// get the "example" plugin, expect it implements network.
	_, err = m.getPluginByName("network", pName)
	if err != ErrNotImplemented {
		t.Fatalf("expect get  ErrNotImplemented, but got %v", err)
	}

	// get the "notExist" plugin.
	_, err = m.getPluginByName("volume", "notExist")
	if err != ErrNotFound {
		t.Fatalf("expect get  ErrNotFound error, but got %v", err)
	}
}

func TestScanPluginDir(t *testing.T) {
	tmpDir, err := setupPluginDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := &pluginManager{
		plugins:         make(map[string]*Plugin),
		pluginSockPaths: []string{tmpDir},
		pluginSpecPaths: []string{tmpDir},
	}

	files := []string{"ultron.spec", "ultron.json", "ceph.sock", "test/test.spec"}

	expectNames := []string{"ultron", "test"}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err := ioutil.WriteFile(path, []byte{}, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	names, err := m.scanPluginDir()
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range names {
		if !utils.StringInSlice(expectNames, name) {
			t.Fatalf("setupPluginDir expect %v, but got %v", expectNames, names)
		}
	}

	for _, name := range expectNames {
		if !utils.StringInSlice(names, name) {
			t.Fatalf("setupPluginDir expect %v, but got %v", expectNames, names)
		}
	}
}

func TestGetAllPlugins(t *testing.T) {
	setupPluginServer()
	defer teardownPluginServer()

	tmpDir, err := setupPluginDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := &pluginManager{
		plugins:         make(map[string]*Plugin),
		pluginSockPaths: []string{tmpDir},
		pluginSpecPaths: []string{tmpDir},
	}

	// plugin name.
	pName := "example"
	// the example plugin implements volume.
	implements := []string{"volume"}
	specPath := path.Join(tmpDir, fmt.Sprintf("%s.spec", pName))

	// write the spec.
	err = ioutil.WriteFile(specPath, []byte(server.URL), 0644)
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(HandShakePath, func(w http.ResponseWriter, r *http.Request) {
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

		// the plugin implement volume plugin
		resp := HandShakeResp{implements}

		content, err := json.Marshal(resp)
		if err != nil {
			t.Fatal(err)
		}

		w.Write(content)
	})

	plugins, err := m.getAllPlugins("volume")
	if err != nil {
		t.Fatal(err)
	}

	if len(plugins) != 1 {
		t.Fatalf("expect to get only one plugin, but got %d plugins", len(plugins))
	}

	if plugins[0].Name != pName {
		t.Fatalf("expect get %s plugin, but got %s plugin", pName, plugins[0].Name)
	}

	if plugins[0].Addr != server.URL {
		t.Fatalf("expect get plugin address %s , but got %s", server.URL, plugins[0].Addr)
	}
}
