package main

import (
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/request"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// APIContainerCreateSuite is the test suite for container create API.
type APIContainerCreateSuite struct{}

func init() {
	check.Suite(&APIContainerCreateSuite{})
}

// SetUpTest does common setup in the beginning of each test.
func (suite *APIContainerCreateSuite) SetUpTest(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	PullImage(c, busyboxImage)
}

// TestBasic test create api is ok.
func (suite *APIContainerCreateSuite) TestBasic(c *check.C) {
	cname := "CreateBasic"

	defaultLxcfsBinds := []string{
		"/var/lib:/var/lib/lxc:shared",
		"/var/lib/lxcfs/proc/uptime:/proc/uptime",
		"/var/lib/lxcfs/proc/swaps:/proc/swaps",
		"/var/lib/lxcfs/proc/stat:/proc/stat",
		"/var/lib/lxcfs/proc/diskstats:/proc/diskstats",
		"/var/lib/lxcfs/proc/meminfo:/proc/meminfo",
		"/var/lib/lxcfs/proc/cpuinfo:/proc/cpuinfo",
	}

	q := url.Values{}
	q.Add("name", cname)

	obj := map[string]interface{}{
		"HostConfig": map[string]interface{}{
			// volume related
			"Binds": []string{
				"/tmp:/tmp",
			},

			// runtime
			// NOTE: please make sure the daemon has added runv
			"Runtime": "runv",

			// policy
			"RestartPolicy": map[string]interface{}{
				"Name": "always",
			},

			// isolation options
			"EnableLxcfs":         true,
			"MemoryWmarkRatio":    int64(30),
			"MemoryExtra":         int64(50),
			"MemoryForceEmptyCtl": 0,
			"ScheLatSwitch":       0,

			// oom setting
			"OomScoreAdj":    100,
			"OomKillDisable": true,
		},
		"Image": busyboxImage,
		"Tty":   true,
	}

	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/create", query, body)
	defer DelContainerForceMultyTime(c, cname)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	// decode response
	{
		got := types.ContainerCreateResp{}
		c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)
		c.Assert(got.ID, check.NotNil)
		c.Assert(got.Name, check.Equals, cname)
	}

	// check inspect result
	{
		resp, err := request.Get("/containers/" + cname + "/json")
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 200)
		got := types.ContainerJSON{}
		c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)

		// isolation related check
		c.Assert(got.HostConfig.EnableLxcfs, check.Equals, true)
		c.Assert(*got.HostConfig.MemoryWmarkRatio, check.Equals, int64(30))
		c.Assert(*got.HostConfig.MemoryExtra, check.Equals, int64(50))
		c.Assert(got.HostConfig.MemoryForceEmptyCtl, check.Equals, int64(0))
		c.Assert(got.HostConfig.ScheLatSwitch, check.Equals, int64(0))

		// runtime check
		c.Assert(got.HostConfig.Runtime, check.Equals, "runv")

		// io related check
		c.Assert(got.Config.Tty, check.Equals, true)

		// oom related check
		c.Assert(got.HostConfig.OomScoreAdj, check.Equals, int64(100))
		c.Assert(*got.HostConfig.OomKillDisable, check.Equals, true)

		// policy
		c.Assert(got.HostConfig.RestartPolicy.Name, check.Equals, "always")
		c.Assert(got.HostConfig.RestartPolicy.MaximumRetryCount, check.Equals, int64(0))

		// volume related
		expectedBinds := []string{"/tmp:/tmp"}
		if environment.IsLxcfsEnabled() {
			expectedBinds = append(expectedBinds, defaultLxcfsBinds...)
		}

		// sort the string before match
		sort.Strings(got.HostConfig.Binds)
		sort.Strings(expectedBinds)
		c.Assert(got.HostConfig.Binds, check.DeepEquals, expectedBinds)
	}
}

// TestBasicWithoutName tests creating container without giving name should succeed.
func (suite *APIContainerCreateSuite) TestBasicWithoutName(c *check.C) {
	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/create", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	// Decode response
	got := types.ContainerCreateResp{}
	c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)
	c.Assert(got.ID, check.NotNil)
	c.Assert(got.Name, check.NotNil)

	DelContainerForceMultyTime(c, got.Name)
}

// TestDupContainer tests create a duplicate container, should return 409.
func (suite *APIContainerCreateSuite) TestDupContainer(c *check.C) {
	cname := "CreateDuplicateContainer"

	q := url.Values{}
	q.Add("name", cname)
	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	query := request.WithQuery(q)
	body := request.WithJSONBody(obj)
	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)
	defer DelContainerForceMultyTime(c, cname)

	// Create a duplicate container
	resp, err = request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 409)
}

// TestNonExistingImg tests using non-existing image return 404.
func (suite *APIContainerCreateSuite) TestNonExistingImg(c *check.C) {
	obj := map[string]interface{}{
		"Image":      "non-existing",
		"HostConfig": map[string]interface{}{},
	}
	body := request.WithJSONBody(obj)

	resp, err := request.Post("/containers/create", body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 404)
}

// TestBadParam tests using bad parameter return 400.
func (suite *APIContainerCreateSuite) TestBadParam(c *check.C) {

	// TODO:
	// 1. Invalid container name, for example length too large, illegal letter.
	// 2. Invalid Parameters
	helpwantedForMissingCase(c, "container api create with bad request")
}

func (suite *APIContainerCreateSuite) TestCreateNvidiaConfig(c *check.C) {
	cname := "TestCreateNvidiaConfig"
	q := url.Values{}
	q.Add("name", cname)
	query := request.WithQuery(q)

	obj := map[string]interface{}{
		"Image": busyboxImage,
		"HostConfig": map[string]interface{}{
			"NvidiaConfig": map[string]interface{}{
				"NvidiaDriverCapabilities": "all",
				"NvidiaVisibleDevices":     "none",
			},
		},
	}
	body := request.WithJSONBody(obj)

	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	resp, err = request.Get("/containers/" + cname + "/json")
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 200)

	got := types.ContainerJSON{}
	err = request.DecodeBody(&got, resp.Body)
	c.Assert(err, check.IsNil)

	c.Assert(got.HostConfig.Resources.NvidiaConfig.NvidiaVisibleDevices, check.Equals, "none")
	c.Assert(got.HostConfig.Resources.NvidiaConfig.NvidiaDriverCapabilities, check.Equals, "all")

	DelContainerForceMultyTime(c, cname)
}

func (suite *APIContainerCreateSuite) TestCreateWithMacAddress(c *check.C) {
	cname := "TestCreateWithMacAddress"
	mac := "00:16:3e:02:00:b7"
	q := url.Values{}
	q.Add("name", cname)
	query := request.WithQuery(q)

	obj := map[string]interface{}{
		"Entrypoint":  []string{"top"},
		"Image":       busyboxImage,
		"MacAddress":  mac,
		"NetworkMode": "bridge",
	}
	body := request.WithJSONBody(obj)

	resp, err := request.Post("/containers/create", query, body)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 201)

	q = url.Values{}
	q.Add("detachKeys", "EOF")
	query = request.WithQuery(q)

	resp, err = request.Post("/containers/"+cname+"/start", query)
	c.Assert(err, check.IsNil)
	CheckRespStatus(c, resp, 204)

	ret := command.PouchRun("exec", cname, "ip", "addr", "show", "eth0")
	ret.Assert(c, icmd.Success)

	found := false
	for _, line := range strings.Split(ret.Stdout(), "\n") {
		if strings.Contains(line, "link/ether") {
			if mac == strings.Fields(line)[1] {
				found = true
				break
			}
		}
	}

	c.Assert(found, check.Equals, true)

	DelContainerForceMultyTime(c, cname)
}

func (suite *APIContainerCreateSuite) TestBasicWithSpecificID(c *check.C) {
	{
		specificID := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
		obj := map[string]interface{}{
			"Image":      busyboxImage,
			"HostConfig": map[string]interface{}{},
			"SpecificID": specificID,
		}

		body := request.WithJSONBody(obj)
		resp, err := request.Post("/containers/create", body)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 201)

		// Decode response
		got := types.ContainerCreateResp{}
		c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)
		c.Assert(got.ID, check.Equals, specificID)
		c.Assert(got.Name, check.NotNil)

		DelContainerForceMultyTime(c, got.ID)
	}

	//transfer specific id by url params
	{
		specificID := "0123456789fedcbaab0123456789abcdef0123456789abcdef0123456789abcd"
		q := url.Values{}
		q.Add("specificId", specificID)
		query := request.WithQuery(q)
		obj := map[string]interface{}{
			"Image":      busyboxImage,
			"HostConfig": map[string]interface{}{},
		}

		body := request.WithJSONBody(obj)
		resp, err := request.Post("/containers/create", query, body)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 201)

		// Decode response
		got := types.ContainerCreateResp{}
		c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)
		c.Assert(got.ID, check.Equals, specificID)
		c.Assert(got.Name, check.NotNil)

		DelContainerForceMultyTime(c, got.ID)
	}
}

func (suite *APIContainerCreateSuite) TestInvalidSpecificID(c *check.C) {
	obj := map[string]interface{}{
		"Image":      busyboxImage,
		"HostConfig": map[string]interface{}{},
	}

	//case1: specificID len is less than 64
	{
		obj["SpecificID"] = "123456789ab"
		body := request.WithJSONBody(obj)
		resp, err := request.Post("/containers/create", body)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, http.StatusBadRequest)
	}

	//case2: specificID len is more than 64
	{
		obj["SpecificID"] = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0"
		body := request.WithJSONBody(obj)
		resp, err := request.Post("/containers/create", body)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, http.StatusBadRequest)
	}

	//case3: characters of specificID is not in "0123456789abcdef"
	{
		obj["SpecificID"] = "0123456789abcdefhi0123456789abcdef0123456789abcdef0123456789abcd"
		body := request.WithJSONBody(obj)
		resp, err := request.Post("/containers/create", body)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, http.StatusBadRequest)
	}

	//case4: specificID is conflict with existing one
	{
		specificID := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
		obj := map[string]interface{}{
			"Image":      busyboxImage,
			"HostConfig": map[string]interface{}{},
			"SpecificID": specificID,
		}

		body := request.WithJSONBody(obj)
		resp, err := request.Post("/containers/create", body)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, 201)

		// Decode response
		got := types.ContainerCreateResp{}
		c.Assert(request.DecodeBody(&got, resp.Body), check.IsNil)
		c.Assert(got.ID, check.Equals, specificID)
		c.Assert(got.Name, check.NotNil)

		defer DelContainerForceMultyTime(c, got.ID)

		//create container with existing id
		resp, err = request.Post("/containers/create", body)
		c.Assert(err, check.IsNil)
		CheckRespStatus(c, resp, http.StatusConflict)
	}
}
