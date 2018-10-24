package plugins

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pouch/pkg/utils"

	"github.com/sirupsen/logrus"
)

// pluginManager is a plugin manager, which manages all plugin events.
type pluginManager struct {
	sync.Mutex
	plugins         map[string]*Plugin
	pluginSockPaths []string
	pluginSpecPaths []string
}

// setPluginSockPaths sets the plugin sock paths.
func (m *pluginManager) setPluginSockPaths(paths []string) {
	m.Lock()
	defer m.Unlock()

	m.pluginSockPaths = paths
}

// setPluginSpecPaths sets the plugin spec paths.
func (m *pluginManager) setPluginSpecPaths(paths []string) {
	m.Lock()
	defer m.Unlock()

	m.pluginSpecPaths = paths
}

// getPluginByName returns the requesed plugin.
func (m *pluginManager) getPluginByName(pluginType, name string) (*Plugin, error) {
	m.Lock()
	existPlugin, ok := m.plugins[name]
	m.Unlock()

	if !ok {
		plugin, err := m.retryLoad(name, true)
		if err != nil {
			return nil, err
		}
		existPlugin = plugin
	}

	if existPlugin.implement(pluginType) {
		return existPlugin, nil
	}

	return nil, ErrNotImplemented
}

// getAllPlugins returns all the plugins implement the given pluginType
func (m *pluginManager) getAllPlugins(pluginType string) ([]*Plugin, error) {
	names, err := m.scanPluginDir()
	if err != nil {
		return nil, err
	}

	var (
		plugins []*Plugin
		lock    sync.Mutex
		wg      sync.WaitGroup
	)

	for i := len(names); i > 0; i-- {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			m.Lock()
			existPlugin, ok := m.plugins[name]
			m.Unlock()

			if !ok {
				plugin, err := m.retryLoad(name, false)
				if err != nil {
					return
				}
				existPlugin = plugin
			}

			if existPlugin.implement(pluginType) {
				lock.Lock()
				plugins = append(plugins, existPlugin)
				lock.Unlock()
			}

		}(names[i-1])
	}

	wg.Wait()

	return plugins, nil
}

// scanPluginDir scans the plugin dir and returns the possible plugin names.
func (m *pluginManager) scanPluginDir() ([]string, error) {
	var names []string

	// scan the sock path.
	for _, root := range m.pluginSockPaths {
		if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			// ignore the dir and none-socket files.
			if err != nil || info.Mode()&os.ModeSocket == 0 {
				return nil
			}

			if filepath.Ext(path) == ".sock" {
				names = append(names, strings.TrimSuffix(info.Name(), filepath.Ext(path)))
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	for _, root := range m.pluginSpecPaths {
		if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			// ignore the dir.
			if err != nil || info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".spec" || filepath.Ext(path) == ".json" {
				names = append(names, strings.TrimSuffix(info.Name(), filepath.Ext(path)))
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	// DeDuplicate.
	return utils.DeDuplicate(names), nil

}

// retryLoad will load the plugin. If failed, it will try to load the plugin again.
func (m *pluginManager) retryLoad(name string, retry bool) (*Plugin, error) {
	var start = time.Now()
	var times = 0

	for {
		plugin, err := m.newPlugin(name)
		if err != nil {
			if retry {
				delay := backoff(times)
				if timeout(start, delay) {
					return nil, err
				}
				times++
				logrus.Warnf("plugin %s not found, retry loading after %d seconds", name, delay/time.Second)
				time.Sleep(delay)
				continue
			}
			return nil, err
		}

		// add to the store.
		m.Lock()
		existPlugin, ok := m.plugins[name]
		if ok {
			plugin = existPlugin
		}
		m.plugins[name] = plugin
		m.Unlock()

		// Probe the plugin.
		err = plugin.probe()
		if err != nil {
			m.Lock()
			delete(m.plugins, name)
			m.Unlock()

			return nil, err
		}

		return plugin, nil
	}
}

// newPlugin creates a plugin by given name.
func (m *pluginManager) newPlugin(name string) (*Plugin, error) {
	// first try to create from sock path.
	var sockPaths []string
	for _, dir := range m.pluginSockPaths {
		sockPaths = append(sockPaths, generatePluginPaths(dir, name, ".sock")...)
	}

	for _, path := range sockPaths {
		fi, err := os.Stat(path)
		if err != nil || fi.Mode()&os.ModeSocket == 0 {
			continue
		}

		addr := "unix://" + path

		return m.newPluginFromAddr(name, addr)
	}

	// then load from spec path.
	var specPaths []string
	for _, dir := range m.pluginSpecPaths {
		// add spec file.
		specPaths = append(specPaths, generatePluginPaths(dir, name, ".spec")...)
		// add json file.
		specPaths = append(specPaths, generatePluginPaths(dir, name, ".json")...)
	}

	for _, path := range specPaths {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			continue
		}

		if strings.HasSuffix(path, ".spec") {
			return m.newPluginFromAddr(name, string(content))
		}
		return m.newPluginFromJSON(name, content)
	}

	return nil, ErrNotFound
}

// newPluginFromAddr creates a plugin from name and address.
func (m *pluginManager) newPluginFromAddr(name, addr string) (*Plugin, error) {
	plugin := &Plugin{
		Name:   name,
		Addr:   addr,
		probed: false,
	}

	client, err := NewPluginClient(addr, nil)
	if err != nil {
		return nil, err
	}

	plugin.client = client

	return plugin, nil
}

// newPluginFromJSON creates a plugin from name and json.
func (m *pluginManager) newPluginFromJSON(name string, jsonContent []byte) (*Plugin, error) {
	plugin := new(Plugin)

	if err := json.Unmarshal(jsonContent, plugin); err != nil {
		return nil, err
	}

	client, err := NewPluginClient(plugin.Addr, plugin.TLSConfig)
	if err != nil {
		return nil, err
	}

	plugin.client = client

	return plugin, nil
}

// generatePluginPaths generates all possible paths.
func generatePluginPaths(dir, name, ext string) []string {
	return []string{
		filepath.Join(dir, name),
		filepath.Join(dir, name+ext),
	}
}
