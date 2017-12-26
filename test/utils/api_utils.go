package utils

import (
	"net/url"

	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// RunBusyboxContainer creates
func RunBusyboxContainer(c *check.C, containerName string) {
	CreateBusyboxContainer(c, containerName)
	StartBusyboxContainer(c, containerName)
}

var busyboxImage = "registry.hub.docker.com/library/busybox:latest"

// CreateBusyboxContainer creates
func CreateBusyboxContainer(c *check.C, containerName string) {
	q := url.Values{}
	q.Add("name", containerName)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"Cmd":        []string{"sleep", "10000"},
		"HostConfig": map[string]interface{}{},
	}

	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)

	resp, err := request.Post(c, "/containers/create", query, body)
	c.Assert(resp.StatusCode, check.Equals, 201, err.Error())
}

// StartBusyboxContainer start
func StartBusyboxContainer(c *check.C, containerName string) {
	resp, err := request.Post(c, "/containers/"+containerName+"/start")
	c.Assert(resp.StatusCode, check.Equals, 204, err.Error())
}
