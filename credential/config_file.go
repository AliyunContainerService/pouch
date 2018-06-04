package credential

import "github.com/alibaba/pouch/apis/types"

var (
	defaultRegistry = "docker.io"
	configFileName  = ".pouch/config.json"
)

// ConfigFile defines configs that file needs keep.
type ConfigFile struct {
	AuthConfigs map[string]types.AuthConfig `json:"auths"`
}
