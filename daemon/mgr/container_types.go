package mgr

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/cri/stream/remotecommand"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/docker/go-connections/nat"
	"github.com/docker/libnetwork"
	"github.com/docker/libnetwork/netlabel"
	"github.com/docker/libnetwork/options"
	networktypes "github.com/docker/libnetwork/types"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// DefaultStopTimeout is the timeout (in seconds) for the syscall signal used to stop a container.
	DefaultStopTimeout = 10
)

var (
	errInvalidEndpoint = "invalid endpoint while building port map info"
	errInvalidNetwork  = "invalid network settings while building port map info"
)

// ContainerFilter defines a function to filter
// container in the store.
type ContainerFilter func(*ContainerMeta) bool

// ContainerExecConfig is the config a process exec.
type ContainerExecConfig struct {
	// ExecID identifies the ID of this exec
	ExecID string

	// contains the config of this exec
	types.ExecCreateConfig

	// Save the container's id into exec config.
	ContainerID string

	// ExitCode records the exit code of a exec process.
	ExitCode int64

	// Running represents whether the exec process is running inside container.
	Running bool

	// Error represents the exec process response error.
	Error error
}

// AttachConfig wraps some infos of attaching.
type AttachConfig struct {
	Stdin  bool
	Stdout bool
	Stderr bool

	// Attach using http.
	Hijack  http.Hijacker
	Upgrade bool

	// Attach using memory buffer.
	MemBuffer *bytes.Buffer

	// Attach using streams.
	Streams *remotecommand.Streams

	// Attach to the container to get its log.
	CriLogFile *os.File
}

// ContainerRemoveOption wraps the container remove interface params.
type ContainerRemoveOption struct {
	Force  bool
	Volume bool
	Link   bool
}

// ContainerListOption wraps the container list interface params.
type ContainerListOption struct {
	All bool
}

// ContainerMeta represents the container's meta data.
type ContainerMeta struct {
	sync.Mutex

	// app armor profile
	AppArmorProfile string `json:"AppArmorProfile,omitempty"`

	// seccomp profile
	SeccompProfile string `json:"SeccompProfile,omitempty"`

	// no new privileges
	NoNewPrivileges bool `json:"NoNewPrivileges,omitempty"`

	// The arguments to the command being run
	Args []string `json:"Args"`

	// config
	Config *types.ContainerConfig `json:"Config,omitempty"`

	// The time the container was created
	Created string `json:"Created,omitempty"`

	// driver
	Driver string `json:"Driver,omitempty"`

	// exec ids
	ExecIds string `json:"ExecIDs,omitempty"`

	// graph driver
	GraphDriver *types.GraphDriverData `json:"GraphDriver,omitempty"`

	// host config
	HostConfig *types.HostConfig `json:"HostConfig,omitempty"`

	// hostname path
	HostnamePath string `json:"HostnamePath,omitempty"`

	// hosts path
	HostsPath string `json:"HostsPath,omitempty"`

	// The ID of the container
	ID string `json:"Id,omitempty"`

	// The container's image
	Image string `json:"Image,omitempty"`

	// log path
	LogPath string `json:"LogPath,omitempty"`

	// mount label
	MountLabel string `json:"MountLabel,omitempty"`

	// mounts
	Mounts []*types.MountPoint `json:"Mounts"`

	// name
	Name string `json:"Name,omitempty"`

	// network settings
	NetworkSettings *types.NetworkSettings `json:"NetworkSettings,omitempty"`

	Node interface{} `json:"Node,omitempty"`

	// The path to the command being run
	Path string `json:"Path,omitempty"`

	// process label
	ProcessLabel string `json:"ProcessLabel,omitempty"`

	// resolv conf path
	ResolvConfPath string `json:"ResolvConfPath,omitempty"`

	// restart count
	RestartCount int64 `json:"RestartCount,omitempty"`

	// The total size of all the files in this container.
	SizeRootFs int64 `json:"SizeRootFs,omitempty"`

	// The size of files that have been created or changed by this container.
	SizeRw int64 `json:"SizeRw,omitempty"`

	// state
	State *types.ContainerState `json:"State,omitempty"`

	// BaseFS
	BaseFS string `json:"BaseFS, omitempty"`
}

// Key returns container's id.
func (meta *ContainerMeta) Key() string {
	return meta.ID
}

func (meta *ContainerMeta) merge(getconfig func() (v1.ImageConfig, error)) error {
	config, err := getconfig()
	if err != nil {
		return err
	}

	// If user specify the Entrypoint, no need to merge image's configuration.
	// Otherwise use the image's configuration to fill it.
	if len(meta.Config.Entrypoint) == 0 {
		if len(meta.Config.Cmd) == 0 {
			meta.Config.Cmd = config.Cmd
		}
		meta.Config.Entrypoint = config.Entrypoint
	}
	if meta.Config.Env == nil {
		meta.Config.Env = config.Env
	} else {
		meta.Config.Env = append(meta.Config.Env, config.Env...)
	}
	if meta.Config.WorkingDir == "" {
		meta.Config.WorkingDir = config.WorkingDir
	}

	return nil
}

// FormatStatus format container status
func (meta *ContainerMeta) FormatStatus() (string, error) {
	var status string

	// return status if container is not running
	if meta.State.Status != types.StatusRunning && meta.State.Status != types.StatusPaused {
		return string(meta.State.Status), nil
	}

	// format container status if container is running
	start, err := time.Parse(utils.TimeLayout, meta.State.StartedAt)
	if err != nil {
		return "", err
	}

	startAt, err := utils.FormatTimeInterval(start.UnixNano())
	if err != nil {
		return "", err
	}

	status = "Up " + startAt
	if meta.State.Status == types.StatusPaused {
		status += "(paused)"
	}
	return status, nil
}

// BuildEndpointInfo sets endpoint-related fields on container.NetworkSettings based on the provided network and endpoint.
func (meta *ContainerMeta) BuildEndpointInfo(n libnetwork.Network, ep libnetwork.Endpoint) error {
	if ep == nil {
		return fmt.Errorf("invalid endpoint while building port map info")
	}

	networkSettings := meta.NetworkSettings
	if networkSettings == nil {
		return fmt.Errorf("invalid network settings while building port map info")
	}

	epInfo := ep.Info()
	if epInfo == nil {
		return nil
	}

	if _, ok := networkSettings.Networks[n.Name()]; !ok {
		networkSettings.Networks[n.Name()] = &types.EndpointSettings{}
	}
	networkSettings.Networks[n.Name()].NetworkID = n.ID()
	networkSettings.Networks[n.Name()].EndpointID = ep.ID()

	iface := epInfo.Iface()
	if iface == nil {
		return nil
	}

	if iface.MacAddress() != nil {
		networkSettings.Networks[n.Name()].MacAddress = iface.MacAddress().String()
	}

	if iface.Address() != nil {
		ones, _ := iface.Address().Mask.Size()
		networkSettings.Networks[n.Name()].IPAddress = iface.Address().IP.String()
		networkSettings.Networks[n.Name()].IPPrefixLen = int64(ones)
	}

	if iface.AddressIPv6() != nil && iface.AddressIPv6().IP.To16() != nil {
		onesv6, _ := iface.AddressIPv6().Mask.Size()
		networkSettings.Networks[n.Name()].GlobalIPV6Address = iface.AddressIPv6().IP.String()
		networkSettings.Networks[n.Name()].GlobalIPV6PrefixLen = int64(onesv6)
	}

	return nil
}

// UpdateSandboxNetworkSettings updates the sandbox ID and Key.
func (meta *ContainerMeta) UpdateSandboxNetworkSettings(sb libnetwork.Sandbox) error {
	meta.NetworkSettings.SandboxID = sb.ID()
	meta.NetworkSettings.SandboxKey = sb.Key()
	return nil
}

func (meta *ContainerMeta) buildPortMapInfo(ep libnetwork.Endpoint) error {
	if ep == nil {
		return fmt.Errorf(errInvalidEndpoint)
	}

	networkSettings := meta.NetworkSettings
	if networkSettings == nil {
		return fmt.Errorf(errInvalidNetwork)
	}

	if len(networkSettings.Ports) == 0 {
		pm, err := getEndpointPortMapInfo(ep)
		if err != nil {
			return err
		}

		var portMap types.PortMap

		for key, value := range pm {
			for _, element := range value {
				portBinding := types.PortBinding{
					HostIP:   element.HostIP,
					HostPort: element.HostPort,
				}
				portMap[string(key)] = append(portMap[string(key)], portBinding)
			}
		}
		networkSettings.Ports = portMap
	}
	return nil
}

// UpdateJoinInfo updates network settings when container joins network n with endpoint ep.
func (meta *ContainerMeta) UpdateJoinInfo(name string, ep libnetwork.Endpoint) error {
	if err := meta.buildPortMapInfo(ep); err != nil {
		return err
	}

	epInfo := ep.Info()
	if epInfo == nil {
		// It is not an error to get an empty endpoint info
		return nil
	}
	if epInfo.Gateway() != nil {
		meta.NetworkSettings.Networks[name].Gateway = epInfo.Gateway().String()
	}
	if epInfo.GatewayIPv6().To16() != nil {
		meta.NetworkSettings.Networks[name].IPV6Gateway = epInfo.GatewayIPv6().String()
	}

	return nil
}

// BuildCreateEndpointOptions builds endpoint options from a given network.
func (meta *ContainerMeta) BuildCreateEndpointOptions(n libnetwork.Network, epConfig *types.EndpointSettings, sb libnetwork.Sandbox) ([]libnetwork.EndpointOption, error) {
	var (
		bindings      = make(nat.PortMap)
		pbList        []networktypes.PortBinding
		exposeList    []networktypes.TransportPort
		createOptions []libnetwork.EndpointOption
	)

	// TODO not sure if defaultNetName is given the right value.
	defaultNetName := NetworkName(meta.HostConfig.NetworkMode)

	if epConfig != nil {
		ipam := epConfig.IPAMConfig

		if ipam != nil {
			var (
				ipList          []net.IP
				ip, ip6, linkip net.IP
			)
			for _, ips := range ipam.LinkLocalIps {
				if linkip = net.ParseIP(ips); linkip == nil && ips != "" {
					return nil, fmt.Errorf("Invalid link-local IP address: %s", ipam.LinkLocalIps)
				}
				ipList = append(ipList, linkip)
			}

			if ip = net.ParseIP(ipam.IPV4Address); ip == nil && ipam.IPV4Address != "" {
				return nil, fmt.Errorf("Invalid IPv4 address: %s)", ipam.IPV4Address)
			}

			if ip6 = net.ParseIP(ipam.IPV6Address); ip6 == nil && ipam.IPV6Address != "" {
				return nil, fmt.Errorf("Invalid IPv6 address: %s)", ipam.IPV6Address)
			}

			createOptions = append(createOptions,
				libnetwork.CreateOptionIpam(ip, ip6, ipList, nil))

		}

		for _, alias := range epConfig.Aliases {
			createOptions = append(createOptions, libnetwork.CreateOptionMyAlias(alias))
		}
		for k, v := range epConfig.DriverOpts {
			createOptions = append(createOptions, libnetwork.EndpointOptionGeneric(options.Generic{k: v}))
		}
	}

	if IsUserDefined(n.Name()) {
		createOptions = append(createOptions, libnetwork.CreateOptionDisableResolution())
	}

	if n.Name() == NetworkName(meta.HostConfig.NetworkMode) ||
		(n.Name() == defaultNetName && IsDefault(meta.HostConfig.NetworkMode)) {
		if meta.Config.MacAddress != "" {
			mac, err := net.ParseMAC(meta.Config.MacAddress)
			if err != nil {
				return nil, err
			}

			genericOption := options.Generic{
				netlabel.MacAddress: mac,
			}

			createOptions = append(createOptions, libnetwork.EndpointOptionGeneric(genericOption))
		}

	}

	// Port-mapping rules belong to the container & applicable only to non-internal networks
	portmaps := GetSandboxPortMapInfo(sb)
	if n.Info().Internal() || len(portmaps) > 0 {
		return createOptions, nil
	}

	if meta.HostConfig.PortBindings != nil {
		for p, b := range meta.HostConfig.PortBindings {
			bindings[nat.Port(p)] = []nat.PortBinding{}
			for _, bb := range b {
				bindings[nat.Port(p)] = append(bindings[nat.Port(p)], nat.PortBinding{
					HostIP:   bb.HostIP,
					HostPort: bb.HostPort,
				})
			}
		}
	}

	portSpecs := meta.Config.ExposedPorts
	ports := make([]nat.Port, len(portSpecs))
	var i int
	for p := range portSpecs {
		ports[i] = nat.Port(p)
		i++
	}
	nat.SortPortMap(ports, bindings)
	for _, port := range ports {
		expose := networktypes.TransportPort{}
		expose.Proto = networktypes.ParseProtocol(port.Proto())
		expose.Port = uint16(port.Int())
		exposeList = append(exposeList, expose)

		pb := networktypes.PortBinding{Port: expose.Port, Proto: expose.Proto}
		binding := bindings[port]
		for i := 0; i < len(binding); i++ {
			pbCopy := pb.GetCopy()
			newP, err := nat.NewPort(nat.SplitProtoPort(binding[i].HostPort))
			var portStart, portEnd int
			if err == nil {
				portStart, portEnd, err = newP.Range()
			}
			if err != nil {
				return nil, err
			}
			pbCopy.HostPort = uint16(portStart)
			pbCopy.HostPortEnd = uint16(portEnd)
			pbCopy.HostIP = net.ParseIP(binding[i].HostIP)
			pbList = append(pbList, pbCopy)
		}

		if meta.HostConfig.PublishAllPorts && len(binding) == 0 {
			pbList = append(pbList, pb)
		}
	}

	createOptions = append(createOptions,
		libnetwork.CreateOptionPortMapping(pbList),
		libnetwork.CreateOptionExposedPorts(exposeList))

	return createOptions, nil
}

// BuildJoinOptions builds endpoint Join options from a given network.
func (meta *ContainerMeta) BuildJoinOptions(n libnetwork.Network) ([]libnetwork.EndpointOption, error) {
	var joinOptions []libnetwork.EndpointOption
	if epConfig, ok := meta.NetworkSettings.Networks[n.Name()]; ok {
		for _, str := range epConfig.Links {
			name, alias, err := ParseLink(str)
			if err != nil {
				return nil, err
			}
			joinOptions = append(joinOptions, libnetwork.CreateOptionAlias(name, alias))
		}
	}
	return joinOptions, nil
}

func getEndpointPortMapInfo(ep libnetwork.Endpoint) (nat.PortMap, error) {
	pm := nat.PortMap{}
	driverInfo, err := ep.DriverInfo()
	if err != nil {
		return pm, err
	}

	if driverInfo == nil {
		// It is not an error for epInfo to be nil
		return pm, nil
	}

	if expData, ok := driverInfo[netlabel.ExposedPorts]; ok {
		if exposedPorts, ok := expData.([]networktypes.TransportPort); ok {
			for _, tp := range exposedPorts {
				natPort, err := nat.NewPort(tp.Proto.String(), strconv.Itoa(int(tp.Port)))
				if err != nil {
					return pm, fmt.Errorf("Error parsing Port value(%v):%v", tp.Port, err)
				}
				pm[natPort] = nil
			}
		}
	}

	mapData, ok := driverInfo[netlabel.PortMap]
	if !ok {
		return pm, nil
	}

	if portMapping, ok := mapData.([]networktypes.PortBinding); ok {
		for _, pp := range portMapping {
			natPort, err := nat.NewPort(pp.Proto.String(), strconv.Itoa(int(pp.Port)))
			if err != nil {
				return pm, err
			}
			natBndg := nat.PortBinding{HostIP: pp.HostIP.String(), HostPort: strconv.Itoa(int(pp.HostPort))}
			pm[natPort] = append(pm[natPort], natBndg)
		}
	}

	return pm, nil
}

// GetSandboxPortMapInfo retrieves the current port-mapping programmed for the given sandbox
func GetSandboxPortMapInfo(sb libnetwork.Sandbox) nat.PortMap {
	pm := nat.PortMap{}
	if sb == nil {
		return pm
	}

	for _, ep := range sb.Endpoints() {
		pm, _ = getEndpointPortMapInfo(ep)
		if len(pm) > 0 {
			break
		}
	}
	return pm
}

// Container represents the container instance in runtime.
type Container struct {
	sync.Mutex
	meta       *ContainerMeta
	DetachKeys string
}

// Key returns container's id.
func (c *Container) Key() string {
	return c.meta.ID
}

// ID returns container's id.
func (c *Container) ID() string {
	return c.meta.ID
}

// Image returns container's image name.
func (c *Container) Image() string {
	return c.meta.Config.Image
}

// Name returns container's name.
func (c *Container) Name() string {
	return c.meta.Name
}

// Config returns container's config.
func (c *Container) Config() *types.ContainerConfig {
	return c.meta.Config
}

// HostConfig returns container's hostconfig.
func (c *Container) HostConfig() *types.HostConfig {
	return c.meta.HostConfig
}

// IsRunning returns container is running or not.
func (c *Container) IsRunning() bool {
	return c.meta.State.Status == types.StatusRunning
}

// IsStopped returns container is stopped or not.
func (c *Container) IsStopped() bool {
	return c.meta.State.Status == types.StatusStopped
}

// IsExited returns container is exited or not.
func (c *Container) IsExited() bool {
	return c.meta.State.Status == types.StatusExited
}

// IsCreated returns container is created or not.
func (c *Container) IsCreated() bool {
	return c.meta.State.Status == types.StatusCreated
}

// IsPaused returns container is paused or not.
func (c *Container) IsPaused() bool {
	return c.meta.State.Status == types.StatusPaused
}

// IsRestarting returns container is restarting or not.
func (c *Container) IsRestarting() bool {
	return c.meta.State.Status == types.StatusRestarting
}

// Write writes container's meta data into meta store.
func (c *Container) Write(store *meta.Store) error {
	return store.Put(c.meta)
}

// StopTimeout returns the timeout (in seconds) used to stop the container.
func (c *Container) StopTimeout() int64 {
	if c.meta.Config.StopTimeout != nil {
		return *c.meta.Config.StopTimeout
	}
	return DefaultStopTimeout
}

// ContainerRestartPolicy represents the policy is used to manage container.
type ContainerRestartPolicy types.RestartPolicy

// IsNone returns the container don't need to be restarted or not.
func (p ContainerRestartPolicy) IsNone() bool {
	return p.Name == "" || p.Name == "no"
}

// IsAlways returns the container need to be restarted or not.
func (p ContainerRestartPolicy) IsAlways() bool {
	return p.Name == "always"
}
