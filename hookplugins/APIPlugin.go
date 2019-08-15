package hookplugins

import (
	"context"

	"github.com/alibaba/pouch/apis/server/types"
)

// APIPlugin provides the ability to extend PouchContainer HTTP API and change how handler behave.
type APIPlugin interface {
	// UpdateHandler could register extra HTTP API to PouchContainer server,
	// change the behavior of the default handler.
	// The default handler of each API would be passed in while starting HTTP server.
	UpdateHandler(context.Context, []*types.HandlerSpec) []*types.HandlerSpec
}

var apiPlugin APIPlugin

// RegisterAPIPlugin is used to register api container.
func RegisterAPIPlugin(ap APIPlugin) {
	apiPlugin = ap
}

// GetAPIPlugin returns the api plugin.
func GetAPIPlugin() APIPlugin {
	return apiPlugin
}
