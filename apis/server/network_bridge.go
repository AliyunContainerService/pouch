package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
)

func (s *Server) createNetwork(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	var networkCreateReq types.NetworkCreateRequest

	if err := json.NewDecoder(req.Body).Decode(&networkCreateReq); err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return err
	}

	network, err := s.NetworkMgr.NetworkCreate(ctx, networkCreateReq)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}

	networkCreateResp := types.NetworkCreateResponse{
		ID: network.ID,
	}

	resp.WriteHeader(http.StatusCreated)
	return json.NewEncoder(resp).Encode(networkCreateResp)
}

func (s *Server) deleteNetwork(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	return nil
}
