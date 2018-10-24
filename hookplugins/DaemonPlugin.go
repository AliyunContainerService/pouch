package hookplugins

// DaemonPlugin defines places where a plugin will be triggered in pouchd lifecycle
type DaemonPlugin interface {
	// PreStartHook is invoked by pouch daemon before real start, in this hook user could start http proxy or other
	// standalone process plugins
	PreStartHook() error

	// PreStopHook is invoked by pouch daemon before daemon process exit, not a promise if daemon is killed, in this
	// hook user could stop the process or plugin started by PreStartHook
	PreStopHook() error
}

var daemonPlugin DaemonPlugin

// RegisterDaemonPlugin is used to register daemon plugin.
func RegisterDaemonPlugin(dp DaemonPlugin) {
	daemonPlugin = dp
}

// GetDaemonPlugin returns the daemon plugin.
func GetDaemonPlugin() DaemonPlugin {
	return daemonPlugin
}
