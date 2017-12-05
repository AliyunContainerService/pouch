package mgr

import (
	"fmt"
	"context"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/docker/libnetwork"
)

const (
	// DefaultBridgeName is the default name for the bridge interface managed
	// by the driver when unspecified by the caller.
	DefaultBridgeName = "pouch0"
)

// NetworkMgr as an interface defines all operations against networks.
type NetworkMgr interface {
	// ListNetworks lists networks
	ListNetworks(ctx context.Context) ([]types.Network, error) 	
}

// NetworkManager is an implementation of interface NetworkMgr.
// It is a stateless manager, and it will never store network details.
// When network details needed from users, NetworkManager use libnetwork.Controller
// to get details.
type NetworkManager struct {
	controller 		libnetwork.NetworkController
}

// NewNetworkManager initializes a brand new network manager.
func NewNetworkManager(cfg *config.Config) (*NetworkManager, error) {
	controller, err := libnetwork.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create network controller: %v", err)
	}

	// Initialize default network on "null".
	if n, _ := controller.NetworkByName("none"); n == nil {
		if _, err := controller.NewNetwork("null", "none", "", libnetwork.NetworkOptionPersist(true)); err != nil {
			return nil, fmt.Errorf("error creating default \"null\" network: %v", err)
		}
	}

	// Initialize default network on "host".
	if n, _ := controller.NetworkByName("host"); n == nil {
		if _, err := controller.NewNetwork("host", "host", "", libnetwork.NetworkOptionPersist(true)); err != nil {
			return nil, fmt.Errorf("error creating default \"host\" network: %v", err)
		}
	}

	return &NetworkManager{
		controller:		controller,
	}, nil
}

// ListNetworks lists networks
func (mgr *NetworkManager) ListNetworks(ctx context.Context) ([]types.Network, error) {
	list := []types.Network{}

	for _, nw := range mgr.controller.Networks() {
			n := buildNetwork(nw)
			list = append(list, *n)
	}

	return list, nil
}

func buildNetwork(nw libnetwork.Network) *types.Network {
	n := &types.Network{}
	if nw == nil {
		return n
	}

	info := nw.Info()
	n.Name = nw.Name()
	n.ID = nw.ID()
	n.Scope = info.Scope()
	n.Driver = nw.Type()

	return n
}
