package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/test/environment"
	"github.com/go-check/check"
)

var (
	// A apiClient is a pouch API client.
	apiClient *client.APIClient
)

// TestMain will do initializes and run all the cases.
func TestMain(m *testing.M) {
	var err error

	apiClient, err = client.NewAPIClient(environment.PouchdAddress, environment.TLSConfig)
	if err != nil {
		fmt.Printf("fail to initializes pouch API client: %s", err.Error())
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// Test is the entrypoint.
func Test(t *testing.T) {
	check.TestingT(t)
}
