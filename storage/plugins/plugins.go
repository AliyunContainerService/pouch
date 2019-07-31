package plugins

var manager = &pluginManager{
	plugins:         make(map[string]*Plugin),
	pluginSockPaths: []string{"/run/pouch/plugins", "/run/docker/plugins"},
	pluginSpecPaths: []string{
		"/etc/pouch/plugins",
		"/var/lib/pouch/plugins",
		"/etc/docker/plugins",
		"/var/lib/docker/plugins"},
}

// Get returns the requested plugin.
func Get(pluginType, name string) (*Plugin, error) {
	return manager.getPluginByName(pluginType, name)
}

// GetAll returns all the plugins which implement the pluginType.
func GetAll(pluginType string) ([]*Plugin, error) {
	return manager.getAllPlugins(pluginType)
}

// SetPluginSockPaths sets the plugin sock paths.
func SetPluginSockPaths(paths []string) {
	manager.setPluginSockPaths(paths)
}

// SetPluginSpecPaths sets the plugin spec paths.
func SetPluginSpecPaths(paths []string) {
	manager.setPluginSpecPaths(paths)
}
