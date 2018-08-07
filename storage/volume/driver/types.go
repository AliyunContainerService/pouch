package driver

type remoteVolumeCreateReq struct {
	Name string            `json:"Name"`
	Opts map[string]string `json:"Opts"`
}

type remoteVolumeCreateResp struct {
	Err string `json:"Err"`
}

type remoteVolumeRemoveReq struct {
	Name string `json:"Name"`
}

type remoteVolumeMountReq struct {
	Name string `json:"Name"`
	ID   string `json:"ID"`
}

type remoteVolumeMountResp struct {
	Mountpoint string `json:"Mountpoint"`
	Err        string `json:"Err"`
}

type remoteVolumePathReq struct {
	Name string `json:"Name"`
}

type remoteVolumePathResp struct {
	Mountpoint string `json:"Mountpoint"`
	Err        string `json:"Err"`
}

type remoteVolumeUnmountReq struct {
	Name string `json:"Name"`
	ID   string `json:"ID"`
}

type remoteVolumeUnmountResp struct {
	Err string `json:"Err"`
}

type remoteVolumeGetReq struct {
	Name string `json:"Name"`
}

type remoteVolumeDesc struct {
	Name       string                 `json:"Name"`
	Mountpoint string                 `json:"Mountpoint"`
	Status     map[string]interface{} `json:"Status"`
}

type remoteVolumeCapability struct {
	Scope string `json:"Scope"`
}

type remoteVolumeGetResp struct {
	Volume *remoteVolumeDesc `json:"Volume"`
	Err    string            `json:"Err"`
}

type remoteVolumeListReq struct {
}

type remoteVolumeListResp struct {
	Volumes []*remoteVolumeDesc `json:"Volumes"`
	Err     string              `json:"Err"`
}

type remoteVolumeCapabilitiesReq struct {
}

type remoteVolumeCapabilitiesResp struct {
	Capabilities *remoteVolumeCapability `json:"Capabilities"`
	Err          string                  `json:"Err"`
}
