package main

import (
	"github.com/alibaba/pouch/test/environment"

	"github.com/go-check/check"
)

// const defines common image name
var (
	busyboxImage                string
	helloworldImage             string
	helloworldImageOnlyRepoName = "hello-world"

	// GateWay test gateway
	GateWay string

	// Subnet test subnet
	Subnet string
)

const (
	busyboxImage125 = "registry.hub.docker.com/library/busybox:1.25"
	testHubAddress  = "registry.hub.docker.com"
	testHubUser     = "pouchcontainertest"
	testHubPasswd   = "pouchcontainertest"
)

func init() {
	// Get test images config from test environment.
	environment.GetBusybox()
	environment.GetHelloWorld()

	busyboxImage = environment.BusyboxRepo + ":" + environment.BusyboxTag
	helloworldImage = environment.HelloworldRepo + ":" + environment.HelloworldTag

	GateWay = environment.GateWay
	Subnet = environment.Subnet

}

// VerifyCondition is used to check the condition value.
type VerifyCondition func() bool

// SkipIfFalse skips the suite, if any of the conditions is not satisfied.
func SkipIfFalse(c *check.C, conditions ...VerifyCondition) {
	for _, con := range conditions {
		if con() == false {
			c.Skip("Skip test as condition is not matched")
		}
	}
}
