package daemonplugin

import "github.com/alibaba/pouch/hookplugins"

type daemonPlugin struct{}

func init() {
	hookplugins.RegisterDaemonPlugin(&daemonPlugin{})
}

// PreStartHook is invoked by pouch daemon before real start, in this hook user could start http proxy or other
// standalone process plugins
func (d *daemonPlugin) PreStartHook() error {
	// TODO: Implemented by the developer
	return nil
}

// PreStopHook stops plugin processes than start ed by PreStartHook.
func (d *daemonPlugin) PreStopHook() error {
	// TODO: Implemented by the developer
	return nil
}
