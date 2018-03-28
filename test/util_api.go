package main

import (
	"bufio"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// CheckRespStatus checks the http.Response.Status is equal to status.
func CheckRespStatus(c *check.C, resp *http.Response, status int) {
	if resp.StatusCode != status {
		got := types.Error{}
		_ = request.DecodeBody(&got, resp.Body)
		c.Assert(resp.StatusCode, check.Equals, status, check.Commentf("Error:%s", got.Message))
	}
}

// CreateBusyboxContainerOk creates a busybox container and asserts success.
func CreateBusyboxContainerOk(c *check.C, cname string, cmd ...string) {
	// If not specified, CMD executed in container is "top".
	if len(cmd) == 0 {
		cmd = []string{"top"}
	}

	resp, err := CreateBusyboxContainer(c, cname, cmd...)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)
}

// CreateBusyboxContainer creates a basic container using busybox image.
func CreateBusyboxContainer(c *check.C, cname string, cmd ...string) (*http.Response, error) {
	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"Cmd":        cmd,
		"HostConfig": map[string]interface{}{},
	}

	path := "/containers/create"
	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	return request.Post(path, query, body)
}

// StartContainerOk starts the container and asserts success.
func StartContainerOk(c *check.C, cname string) {
	resp, err := StartContainer(c, cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 204)
}

// StartContainer starts the container.
func StartContainer(c *check.C, cname string) (*http.Response, error) {
	return request.Post("/containers/" + cname + "/start")
}

// DelContainerForceOk forcely deletes the container and asserts success.
func DelContainerForceOk(c *check.C, cname string) {
	resp, err := DelContainerForce(c, cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 204)
}

// DelContainerForce forcely deletes the container.
func DelContainerForce(c *check.C, cname string) (*http.Response, error) {
	q := url.Values{}
	q.Add("force", "true")
	return request.Delete("/containers/"+cname, request.WithQuery(q))
}

// DelContainerForceMultyTime forcely deletes the container multy times.
func DelContainerForceMultyTime(c *check.C, cname string) {
	q := url.Values{}
	q.Add("force", "true")

	ch := make(chan bool, 1)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(3 * time.Second)
		done <- true
	}()

	for {
		select {
		case <-ch:
			return
		case <-done:
			return
		case <-ticker.C:
			resp, _ := DelContainerForce(c, cname)
			if resp.StatusCode == 204 {
				ch <- true
			}
		}
	}
}

// StopContainerOk stops the container and asserts success..
func StopContainerOk(c *check.C, cname string) {
	resp, err := StopContainer(c, cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 204)
}

// StopContainer stops the container.
func StopContainer(c *check.C, cname string) (*http.Response, error) {
	return request.Post("/containers/" + cname + "/stop")
}

// PauseContainerOk pauses the container and asserts success..
func PauseContainerOk(c *check.C, cname string) {
	resp, err := PauseContainer(c, cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 204)
}

// PauseContainer pauses the container.
func PauseContainer(c *check.C, cname string) (*http.Response, error) {
	return request.Post("/containers/" + cname + "/pause")
}

// UnpauseContainerOk unpauses the container and asserts success..
func UnpauseContainerOk(c *check.C, cname string) {
	resp, err := UnpauseContainer(c, cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 204)
}

// UnpauseContainer unpauses the container.
func UnpauseContainer(c *check.C, cname string) (*http.Response, error) {
	return request.Post("/containers/" + cname + "/unpause")
}

// IsContainerCreated returns true is container's state is created.
func IsContainerCreated(c *check.C, cname string) (bool, error) {
	return isContainerStateEqual(c, cname, "created")
}

// IsContainerRunning returns true is container's state is running.
func IsContainerRunning(c *check.C, cname string) (bool, error) {
	return isContainerStateEqual(c, cname, "running")
}

func isContainerStateEqual(c *check.C, cname string, status string) (bool, error) {
	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	defer resp.Body.Close()
	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	if got.State == nil {
		return false, nil
	}

	return string(got.State.Status) == status, nil
}

// CreateExecEchoOk exec process's environment with "echo" CMD.
func CreateExecEchoOk(c *check.C, cname string) string {
	// NOTICE:
	// All files in the obj is needed, or start a new process may hang.
	obj := map[string]interface{}{
		"Cmd":          []string{"echo", "test"},
		"Detach":       true,
		"AttachStderr": true,
		"AttachStdout": true,
		"AttachStdin":  true,
		"Privileged":   false,
		"User":         "",
	}
	body := request.WithJSONBody(obj)

	resp, err := request.Post("/containers/"+cname+"/exec", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	var got types.ExecCreateResp
	request.DecodeBody(&got, resp.Body)
	return got.ID
}

// StartContainerExec starts executing a process in the container.
func StartContainerExec(c *check.C, execid string, tty bool, detach bool) (*http.Response, net.Conn, *bufio.Reader, error) {
	obj := map[string]interface{}{
		"Detach": detach,
		"Tty":    tty,
	}

	return request.Hijack("/exec/"+execid+"/start",
		request.WithHeader("Connection", "Upgrade"),
		request.WithHeader("Upgrade", "tcp"),
		request.WithJSONBody(obj))
}

// CreateVolume creates a volume in pouchd.
func CreateVolume(c *check.C, name, driver string) error {
	obj := map[string]interface{}{
		"Driver": driver,
		"Name":   name,
	}
	path := "/volumes/create"
	body := request.WithJSONBody(obj)

	resp, err := request.Post(path, body)
	defer resp.Body.Close()

	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	return err
}

// RemoveVolume removes a volume in pouchd.
func RemoveVolume(c *check.C, name string) error {
	path := "/volumes/" + name
	resp, err := request.Delete(path)
	defer resp.Body.Close()

	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	return err
}

// DelNetworkOk deletes the network and asserts success.
func DelNetworkOk(c *check.C, cname string) {
	resp, err := DelNetwork(c, cname)
	c.Assert(err, check.IsNil)

	CheckRespStatus(c, resp, 204)
}

// DelNetwork  deletes the network.
func DelNetwork(c *check.C, cname string) (*http.Response, error) {
	return request.Delete("/networks/" + cname)
}

// PullImage pull image if it doesn't exist, image format should be repo:tag.
func PullImage(c *check.C, image string) {
	resp, err := request.Get("/images/" + image + "/json")
	c.Assert(err, check.IsNil)

	if resp.StatusCode == 404 {
		q := url.Values{}
		q.Add("fromImage", image)
		query := request.WithQuery(q)
		resp, err = request.Post("/images/create", query)
		c.Assert(resp.StatusCode, check.Equals, 200)
	}
}
