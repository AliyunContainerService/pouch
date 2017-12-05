package server

import (
	"context"
	"net/http"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

func (s *Server) listNetworks(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	networkList, err := s.NetworkMgr.ListNetworks(ctx)
	if err != nil {
		logrus.Errorf("failed to list networks: %v", err)
		return err
	}
	return json.NewEncoder(resp).Encode(networkList)
}
