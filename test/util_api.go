package main

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/jsonstream"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
)

// CheckRespStatus checks the http.Response.Status is equal to status.
func CheckRespStatus(c *check.C, resp *http.Response, status int) {
	if resp.StatusCode != status {
		body, err := ioutil.ReadAll(resp.Body)
		c.Assert(err, check.IsNil)
		c.Assert(resp.StatusCode, check.Equals, status, check.Commentf("Response Body: %v", string(body)))
	}
}

// CreateBusyboxContainerOk creates a busybox container with cmd and asserts OK.
//
// NOTE: If not specified, CMD executed in container is "top".
func CreateBusyboxContainerOk(c *check.C, cname string, cmd ...string) string {
	if len(cmd) == 0 {
		cmd = []string{"top"}
	}

	resp, err := CreateBusyboxContainer(cname, cmd...)
	c.Assert(err, check.IsNil)

	defer resp.Body.Close()
	CheckRespStatus(c, resp, 201)

	got := types.ContainerCreateResp{}
	c.Assert(json.NewDecoder(resp.Body).Decode(&got), check.IsNil)
	return got.ID
}

// CreateBusyboxContainer creates busybox with cmd.
func CreateBusyboxContainer(cname string, cmd ...string) (*http.Response, error) {
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

// CreateBusybox125Container creates busybox with cmd.
func CreateBusybox125Container(cname string, cmd ...string) (*http.Response, error) {
	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"Image":      busyboxImage125,
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
	resp, err := request.Post("/containers/" + cname + "/start")
	c.Assert(err, check.IsNil)

	defer resp.Body.Close()
	CheckRespStatus(c, resp, 204)
}

// StopContainerOk stops the container and asserts success..
func StopContainerOk(c *check.C, cname string) {
	q := url.Values{}
	q.Add("t", "1")
	query := request.WithQuery(q)

	resp, err := request.Post("/containers/"+cname+"/stop", query)
	c.Assert(err, check.IsNil)

	defer resp.Body.Close()
	CheckRespStatus(c, resp, 204)
}

// CheckContainerStatus asserts the container status.
func CheckContainerStatus(c *check.C, cname string, state string) {
	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)
	defer resp.Body.Close()
	c.Assert(string(got.State.Status), check.Equals, state)
}

// CheckContainerRunning checks if container is running.
func CheckContainerRunning(c *check.C, cname string, isRunning bool) {
	resp, err := request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)
	defer resp.Body.Close()
	gotRunning := (string(got.State.Status) == "running")
	c.Assert(gotRunning, check.Equals, isRunning)
}

func delContainerForce(cname string) (*http.Response, error) {
	// first stop the container, then delete it
	q := url.Values{}
	q.Add("t", "1")
	resp, _ := request.Post("/containers/"+cname+"/stop", request.WithQuery(q))
	defer resp.Body.Close()

	q = url.Values{}
	q.Add("force", "true")
	q.Add("v", "true")

	return request.Delete("/containers/"+cname, request.WithQuery(q))
}

// PauseContainerOk pauses the container and asserts success..
func PauseContainerOk(c *check.C, cname string) {
	resp, err := request.Post("/containers/" + cname + "/pause")
	c.Assert(err, check.IsNil)

	defer resp.Body.Close()
	CheckRespStatus(c, resp, 204)
}

// UnpauseContainerOk unpauses the container and asserts success..
func UnpauseContainerOk(c *check.C, cname string) {
	resp, err := request.Post("/containers/" + cname + "/unpause")
	c.Assert(err, check.IsNil)
	defer resp.Body.Close()
	CheckRespStatus(c, resp, 204)
}

// DelContainerForceMultyTime forcely deletes the container multy times.
func DelContainerForceMultyTime(c *check.C, cname string) {
	timeout := 1 * time.Minute
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			c.Logf("failed to force remove container(%s) in (%s), maybe impact other cases", cname, timeout)
			return
		case <-ticker.C:
			resp, _ := delContainerForce(cname)
			if resp != nil {
				resp.Body.Close()

				if resp.StatusCode == 204 || resp.StatusCode == 404 {
					return
				}
			}
		}
	}
}

// DelImageForceOk forcely deletes the image and asserts success.
func DelImageForceOk(c *check.C, iname string) {
	q := url.Values{}
	q.Add("force", "true")

	resp, err := request.Delete("/images/"+iname, request.WithQuery(q))
	c.Assert(err, check.IsNil)

	defer resp.Body.Close()
	CheckRespStatus(c, resp, 204)
}

// CreateExecCmdOk exec process's environment with specific CMD.
func CreateExecCmdOk(c *check.C, cname string, cmd ...string) string {
	obj := map[string]interface{}{
		"Cmd":          cmd,
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
	c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)
	return got.ID
}

// StartContainerExecOk starts executing a process in container and check if it is successful.
func StartContainerExecOk(c *check.C, execID string) {
	resp, conn, _, err := StartContainerExec(c, execID, false, false)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 101)
	c.Assert(conn.Close(), check.IsNil)
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

// InspectExecOk inspects an exec of container.
func InspectExecOk(c *check.C, execid string) types.ContainerExecInspect {
	resp, err := request.Get("/exec/" + execid + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	var execInspectResp types.ContainerExecInspect
	err = request.DecodeBody(&execInspectResp, resp.Body)
	c.Assert(err, check.IsNil)
	return execInspectResp
}

// CreateVolumeOK creates a volume in pouchd.
func CreateVolumeOK(c *check.C, name, driver string, options map[string]string) {
	obj := map[string]interface{}{
		"Driver":     driver,
		"Name":       name,
		"DriverOpts": options,
	}
	path := "/volumes/create"
	body := request.WithJSONBody(obj)

	resp, err := request.Post(path, body)
	c.Assert(err, check.IsNil)

	defer resp.Body.Close()
	CheckRespStatus(c, resp, 201)
}

// RemoveVolumeOK removes a volume in pouchd.
func RemoveVolumeOK(c *check.C, name string) {
	path := "/volumes/" + name
	resp, err := request.Delete(path)
	c.Assert(err, check.IsNil)

	defer resp.Body.Close()
	CheckRespStatus(c, resp, 204)
}

// DelNetworkOk deletes the network and asserts success.
func DelNetworkOk(c *check.C, cname string) {
	resp, err := request.Delete("/networks/" + cname)
	c.Assert(err, check.IsNil)

	defer resp.Body.Close()
	CheckRespStatus(c, resp, 204)
}

// PullImage pull image if it doesn't exist, image format should be repo:tag.
func PullImage(c *check.C, image string) {
	resp, err := request.Get("/images/" + image + "/json")
	c.Assert(err, check.IsNil)

	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		return
	}

	q := url.Values{}
	q.Add("fromImage", image)
	resp, err = request.Post("/images/create", request.WithQuery(q))
	c.Assert(err, check.IsNil)

	defer resp.Body.Close()
	c.Assert(resp.StatusCode, check.Equals, 200)
	c.Assert(discardPullStatus(resp.Body), check.IsNil)
}

func discardPullStatus(r io.ReadCloser) error {
	dec := json.NewDecoder(r)
	for {
		var msg jsonstream.JSONMessage
		if err := dec.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if msg.Error != nil {
			return msg.Error
		}
	}
	return nil
}

// GetMetric get metrics from prometheus server, return total count and success count.
func GetMetric(c *check.C, key string, keySuccess string) (int, int) {
	resp, err := request.Get("/metrics")
	c.Assert(err, check.IsNil)
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	value := ""
	valueSuccess := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, key) {
			kv := strings.Split(line, " ")
			if len(kv) == 2 {
				value = kv[1]
			}
		} else if strings.Contains(line, keySuccess) {
			kv := strings.Split(line, " ")
			if len(kv) == 2 {
				valueSuccess = kv[1]
			}
		}
	}

	iCount := 0
	if value != "" {
		iCount, err = strconv.Atoi(value)
		c.Assert(err, check.IsNil)
	}

	iCountSuccess := 0
	if valueSuccess != "" {
		iCountSuccess, err = strconv.Atoi(valueSuccess)
		c.Assert(err, check.IsNil)
	}

	return iCount, iCountSuccess
}
