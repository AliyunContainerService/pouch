package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
	networktypes "github.com/alibaba/pouch/network/types"
	"github.com/alibaba/pouch/pkg/httputils"

	"github.com/docker/libnetwork"
	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *Server) createNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.NetworkCreateConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	network, err := s.NetworkMgr.Create(ctx, *config)
	if err != nil {
		return err
	}

	networkCreateResp := types.NetworkCreateResp{
		ID: network.ID,
	}

	return EncodeResponse(rw, http.StatusCreated, networkCreateResp)
}

func (s *Server) listNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	networks, err := s.NetworkMgr.List(ctx, map[string]string{})
	if err != nil {
		return err
	}

	respNetworks := types.NetworkListResp{Networks: []*types.NetworkInfo{}, Warnings: nil}
	for _, net := range networks {
		respNetworks.Networks = append(respNetworks.Networks, &types.NetworkInfo{
			Name:   net.Name,
			ID:     net.ID,
			Driver: net.Type,
		})
	}
	return EncodeResponse(rw, http.StatusOK, respNetworks)
}

func (s *Server) getNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	network, err := s.NetworkMgr.Get(ctx, name)
	if err != nil {
		return err
	}

	networkResp := buildNetworkInspectResp(network)

	return EncodeResponse(rw, http.StatusOK, networkResp)
}

func (s *Server) deleteNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	name := mux.Vars(req)["name"]

	if err := s.NetworkMgr.Remove(ctx, name); err != nil {
		return err
	}
	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) connectToNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	networkIdOrName := mux.Vars(req)["id"]

	net, err := s.NetworkMgr.Get(ctx, networkIdOrName)
	if err != nil || net == nil {
		return err
	}

	connectConfig := &types.NetworkConnectConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(connectConfig); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := connectConfig.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	container, err := s.ContainerMgr.Get(context.Background(), connectConfig.Container)
	if err != nil || container == nil {
		return errors.Wrap(err, "the container to be connected doesn't exist")
	}

	if err := s.NetworkMgr.Connect(ctx, container, networkIdOrName, connectConfig.EndpointConfig); err != nil {
		return err
	}
	rw.WriteHeader(http.StatusOK)
	return nil
}

func buildNetworkInspectResp(n *networktypes.Network) *types.NetworkInspectResp {
	info := n.Network.Info()
	network := &types.NetworkInspectResp{
		Name:       n.Name,
		ID:         n.Network.ID(),
		Driver:     n.Type,
		EnableIPV6: info.IPv6Enabled(),
		Internal:   info.Internal(),
		Options:    info.DriverOptions(),
		Labels:     info.Labels(),
		IPAM:       &types.IPAM{},
	}
	buildIpamResources(network, info)
	return network
}

func buildIpamResources(r *types.NetworkInspectResp, nwInfo libnetwork.NetworkInfo) {
	id, opts, ipv4conf, ipv6conf := nwInfo.IpamConfig()

	ipv4Info, ipv6Info := nwInfo.IpamInfo()

	r.IPAM.Driver = id

	r.IPAM.Options = opts

	r.IPAM.Config = []types.IPAMConfig{}
	for _, ip4 := range ipv4conf {
		if ip4.PreferredPool == "" {
			continue
		}
		iData := types.IPAMConfig{}
		iData.Subnet = ip4.PreferredPool
		iData.IPRange = ip4.SubPool
		iData.Gateway = ip4.Gateway
		iData.AuxAddress = ip4.AuxAddresses
		r.IPAM.Config = append(r.IPAM.Config, iData)
	}

	if len(r.IPAM.Config) == 0 {
		for _, ip4Info := range ipv4Info {
			iData := types.IPAMConfig{}
			iData.Subnet = ip4Info.IPAMData.Pool.String()
			iData.Gateway = ip4Info.IPAMData.Gateway.String()
			r.IPAM.Config = append(r.IPAM.Config, iData)
		}
	}

	hasIpv6Conf := false
	for _, ip6 := range ipv6conf {
		if ip6.PreferredPool == "" {
			continue
		}
		hasIpv6Conf = true
		iData := types.IPAMConfig{}
		iData.Subnet = ip6.PreferredPool
		iData.IPRange = ip6.SubPool
		iData.Gateway = ip6.Gateway
		iData.AuxAddress = ip6.AuxAddresses
		r.IPAM.Config = append(r.IPAM.Config, iData)
	}

	if !hasIpv6Conf {
		for _, ip6Info := range ipv6Info {
			if ip6Info.IPAMData.Pool == nil {
				continue
			}
			iData := types.IPAMConfig{}
			iData.Subnet = ip6Info.IPAMData.Pool.String()
			iData.Gateway = ip6Info.IPAMData.Gateway.String()
			r.IPAM.Config = append(r.IPAM.Config, iData)
		}
	}
}
