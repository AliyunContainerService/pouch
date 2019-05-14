package formatter

import (
	"fmt"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"
)

// ContainerHeader is the map to show container head
var ContainerHeader = map[string]string{
	"Names":        "Name",
	"ID":           "ID",
	"Status":       "Status",
	"RunningFor":   "Created",
	"Image":        "Image",
	"Runtime":      "Runtime",
	"Command":      "Command",
	"ImageID":      "ImageID",
	"Labels":       "Labels",
	"Mounts":       "Mounts",
	"State":        "State",
	"Ports":        "Ports",
	"Size":         "Size",
	"LocalVolumes": "LocalVolumes",
	"Networks":     "Networks",
	"CreatedAt":    "CreatedAt",
}

// ContainerContext is the map to show container context detail
type ContainerContext map[string]string

// NewContainerContext is to generate a ContainerContext to be show
func NewContainerContext(c *types.Container, flagNoTrunc bool) (containerContext ContainerContext, err error) {
	id := c.ID[:6]
	if flagNoTrunc {
		id = c.ID
	}
	runningFor, err := utils.FormatTimeInterval(c.Created)
	if err != nil {
		return nil, err
	}
	createdAt := time.Unix(0, c.Created).String()
	networks := c.HostConfig.NetworkMode
	ports := PortBindingsToString(c.HostConfig.PortBindings)
	size := SizeToString(c.SizeRw, c.SizeRootFs)
	labels := LabelsToString(c.Labels)
	mount := MountPointToString(c.Mounts)
	localVolume := LocalVolumes(c.Mounts)
	containerContext = ContainerContext{
		"Names":        c.Names[0],
		"ID":           id,
		"Status":       c.Status,
		"RunningFor":   fmt.Sprintf("%s ago", runningFor),
		"Image":        c.Image,
		"Runtime":      c.HostConfig.Runtime,
		"Command":      c.Command,
		"ImageID":      c.ImageID,
		"Labels":       labels,
		"Mounts":       mount,
		"State":        c.State,
		"Ports":        ports,
		"Size":         size,
		"LocalVolumes": localVolume,
		"Networks":     networks,
		"CreatedAt":    createdAt,
	}
	return containerContext, nil
}
