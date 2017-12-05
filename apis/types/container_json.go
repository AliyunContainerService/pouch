package types

// ContainerJSON contains response of Remote API:
// GET "/containers/{name:.*}/json"
type ContainerJSON struct {
	ID              string `json:"Id"`
	Created         string
	Path            string
	Args            []string
	State           *ContainerState
	Image           string
	ResolvConfPath  string
	HostnamePath    string
	HostsPath       string
	LogPath         string
	Name            string
	RestartCount    int
	Driver          string
	MountLabel      string
	ProcessLabel    string
	AppArmorProfile string
	ExecIDs         []string
	HostConfig      *HostConfig
	SizeRw          *int64 `json:",omitempty"`
	SizeRootFs      *int64 `json:",omitempty"`
	HostRootPath    string
}

// TODO
// ContainerJSON is newly used struct along with MountPoint
//type ContainerJSON struct {
//	*ContainerJSONBase
//	Mounts          []MountPoint
//	Config          *container.Config
//	NetworkSettings *NetworkSettings
//}
