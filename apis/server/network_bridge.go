package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/httputils"

	"github.com/go-openapi/strfmt"
)

func (s *Server) createNetwork(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	config := &types.NetworkCreateConfig{}
	// decode request body
	if err := json.NewDecoder(req.Body).Decode(config); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}
	// validate request body
	if err := config.Validate(strfmt.NewFormats()); err != nil {
		return httputils.NewHTTPError(err, http.StatusBadRequest)
	}

	network, err := s.NetworkMgr.NetworkCreate(ctx, *config)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}

	networkCreateResp := types.NetworkCreateResp{
		ID: network.ID,
	}

	resp.WriteHeader(http.StatusCreated)
	return json.NewEncoder(resp).Encode(networkCreateResp)
}

func (s *Server) deleteNetwork(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	return nil
}
