package mgr

import (
	"context"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	"github.com/docker/libnetwork"
	"github.com/sirupsen/logrus"
)

// LogContainerEvent generates an event related to a container with only the default attributes.
func (mgr *ContainerManager) LogContainerEvent(ctx context.Context, container *Container, action string) {
	mgr.LogContainerEventWithAttributes(ctx, container, action, map[string]string{})
}

// LogContainerEventWithAttributes generates an event related to a container with specific given attributes.
func (mgr *ContainerManager) LogContainerEventWithAttributes(ctx context.Context, container *Container, action string, attributes map[string]string) {
	copyAttributes(attributes, container.Config.Labels)
	if container.Config.Image != "" {
		attributes["image"] = container.Config.Image
	}
	attributes["name"] = strings.TrimLeft(container.Name, "/")

	actor := &types.EventsActor{
		ID:         container.ID,
		Attributes: attributes,
	}

	_ = mgr.eventsService.Publish(ctx, action, types.EventTypeContainer, actor)
}

// LogVolumeEvent generates an event related to a volume
func (vm *VolumeManager) LogVolumeEvent(ctx context.Context, volumeID, action string, attributes map[string]string) {
	actor := &types.EventsActor{
		ID:         volumeID,
		Attributes: attributes,
	}

	_ = vm.eventsService.Publish(ctx, action, types.EventTypeVolume, actor)
}

// LogNetworkEvent generates an event related to a network with only the default attributes
func (nm *NetworkManager) LogNetworkEvent(ctx context.Context, nw libnetwork.Network, action string) {
	attributes := map[string]string{}
	attributes["name"] = nw.Name()
	attributes["type"] = nw.Type()
	actor := &types.EventsActor{
		ID:         nw.ID(),
		Attributes: attributes,
	}

	_ = nm.eventsService.Publish(ctx, action, types.EventTypeNetwork, actor)
}

// LogNetworkEventWithAttributes generates an event related to a network with specific given attributes
// Use ContainerManager to publish network event may be a little bit ugly now
func (mgr *ContainerManager) LogNetworkEventWithAttributes(ctx context.Context, nw libnetwork.Network, action string, attributes map[string]string) {
	attributes["name"] = nw.Name()
	attributes["type"] = nw.Type()
	actor := &types.EventsActor{
		ID:         nw.ID(),
		Attributes: attributes,
	}

	_ = mgr.eventsService.Publish(ctx, action, types.EventTypeNetwork, actor)
}

// LogImageEvent generates an event related to an image with only the default attributes
func (mgr *ImageManager) LogImageEvent(ctx context.Context, imageID, refName, action string) {
	mgr.LogImageEventWithAttributes(ctx, imageID, refName, action, map[string]string{})
}

// LogImageEventWithAttributes generates an event related to an image with specific given attributes
func (mgr *ImageManager) LogImageEventWithAttributes(ctx context.Context, imageID, refName, action string, attributes map[string]string) {
	img, err := mgr.GetImage(ctx, imageID)
	if err == nil && img.Config != nil {
		copyAttributes(attributes, img.Config.Labels)
	}

	if refName != "" {
		attributes["Name"] = refName
	}
	actor := &types.EventsActor{
		ID:         imageID,
		Attributes: attributes,
	}

	_ = mgr.eventsService.Publish(ctx, action, types.EventTypeImage, actor)
}

// copyAttributes guarantees that labels are not mutated by event triggers.
func copyAttributes(attributes, labels map[string]string) {
	if labels == nil {
		return
	}
	for k, v := range labels {
		attributes[k] = v
	}
}

// publishContainerdEvent sends containerd events to pouchd event service.
func (mgr *ContainerManager) publishContainerdEvent(ctx context.Context, id, action string, attributes map[string]string) error {
	c, err := mgr.container(id)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	mgr.LogContainerEventWithAttributes(ctx, c, action, attributes)

	return nil
}

// updateContainerState update container's state according to the containerd events.
func (mgr *ContainerManager) updateContainerState(ctx context.Context, id, action string, attributes map[string]string) error {
	c, err := mgr.container(id)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	// NOTE: updateContainerState is async-op and it will be executed with
	// ContainerMgmt.Remove at the same time. mgr.container will use
	// memory cache(id <-> container mapping) to get container information.
	// If the updateContainerState gets the container information before
	// ContainerMgmt.Remove but the ContainerMgmt.Remove locks first, the
	// updateContainerState will put the container information into
	// local disk again. Therefore, we should check the status before update
	if c.IsDead() {
		logrus.Warnf("container(%v) is marked dead. no need to update state by action %v", id, action)
		return nil
	}

	dirty := true
	switch action {
	case "oom":
		c.SetStatusOOM()
	default:
		dirty = false
	}

	if dirty {
		if err := mgr.Store.Put(c); err != nil {
			return err
		}
	}

	return nil
}
