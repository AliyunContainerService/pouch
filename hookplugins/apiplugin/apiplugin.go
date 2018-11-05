package apiplugin

import (
	"github.com/alibaba/pouch/apis/server/types"
	"github.com/alibaba/pouch/hookplugins"
)

type apiPlugin struct{}

func init() {
	hookplugins.RegisterAPIPlugin(&apiPlugin{})
}

func (a *apiPlugin) UpdateHandler(handlers []*types.HandlerSpec) []*types.HandlerSpec {
	// TODO: Implemented by the developer
	// just return the original handlers here
	return handlers
}
