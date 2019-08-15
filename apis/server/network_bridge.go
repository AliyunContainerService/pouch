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
)

func (s *Server) createNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	config := &types.NetworkCreateConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	logCreateOptions(ctx, "network", config)

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

func (s *Server) getNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	id := mux.Vars(req)["id"]

	network, err := s.NetworkMgr.Get(ctx, id)
	if err != nil {
		return err
	}

	networkResp := buildNetworkInspectResp(network)

	return EncodeResponse(rw, http.StatusOK, networkResp)
}

func (s *Server) listNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	networks, err := s.NetworkMgr.List(ctx, map[string]string{})
	if err != nil {
		return err
	}

	respNetworks := []types.NetworkResource{}
	for _, net := range networks {
		respNetworks = append(respNetworks, buildNetworkResource(net))
	}
	return EncodeResponse(rw, http.StatusOK, respNetworks)
}

func (s *Server) deleteNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	id := mux.Vars(req)["id"]

	if err := s.NetworkMgr.Remove(ctx, id); err != nil {
		return err
	}
	rw.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) connectToNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	networkIDOrName := mux.Vars(req)["id"]
	connectConfig := &types.NetworkConnect{}

	// decode request body
	if err := json.NewDecoder(req.Body).Decode(connectConfig); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := connectConfig.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	if err := s.ContainerMgr.Connect(ctx, connectConfig.Container, networkIDOrName, connectConfig.EndpointConfig); err != nil {
		return err
	}
	rw.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) disconnectNetwork(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	network := &types.NetworkDisconnect{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(network); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := network.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	id := mux.Vars(req)["id"]

	return s.ContainerMgr.Disconnect(ctx, network.Container, id, network.Force)
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
		IPAM:       buildIpamResources(info),
		Scope:      info.Scope(),
	}
	return network
}

func buildNetworkResource(n *networktypes.Network) types.NetworkResource {
	r := types.NetworkResource{}
	if n == nil {
		return r
	}

	info := n.Network.Info()
	r.Name = n.Name
	r.ID = n.ID
	r.Driver = n.Type
	r.Containers = make(map[string]types.EndpointResource)
	r.EnableIPV6 = info.IPv6Enabled()
	r.IPAM = buildIpamResources(info)
	r.Internal = info.Internal()
	r.Labels = info.Labels()
	r.Options = info.DriverOptions()
	r.Scope = info.Scope()

	return r
}

func buildIpamResources(nwInfo libnetwork.NetworkInfo) *types.IPAM {
	r := &types.IPAM{}

	id, opts, ipv4conf, ipv6conf := nwInfo.IpamConfig()

	ipv4Info, ipv6Info := nwInfo.IpamInfo()

	r.Driver = id

	r.Options = opts

	r.Config = []types.IPAMConfig{}
	for _, ip4 := range ipv4conf {
		if ip4.PreferredPool == "" {
			continue
		}
		iData := types.IPAMConfig{}
		iData.Subnet = ip4.PreferredPool
		iData.IPRange = ip4.SubPool
		iData.Gateway = ip4.Gateway
		iData.AuxAddress = ip4.AuxAddresses
		r.Config = append(r.Config, iData)
	}

	if len(r.Config) == 0 {
		for _, ip4Info := range ipv4Info {
			iData := types.IPAMConfig{}
			iData.Subnet = ip4Info.IPAMData.Pool.String()
			iData.Gateway = ip4Info.IPAMData.Gateway.String()
			r.Config = append(r.Config, iData)
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
		r.Config = append(r.Config, iData)
	}

	if !hasIpv6Conf {
		for _, ip6Info := range ipv6Info {
			if ip6Info.IPAMData.Pool == nil {
				continue
			}
			iData := types.IPAMConfig{}
			iData.Subnet = ip6Info.IPAMData.Pool.String()
			iData.Gateway = ip6Info.IPAMData.Gateway.String()
			r.Config = append(r.Config, iData)
		}
	}

	return r
}
