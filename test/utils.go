package main

import (
	"github.com/go-check/check"
)

// const defines common image name
const (
	busyboxImage                = "registry.hub.docker.com/library/busybox:latest"
	busyboxImage125             = "registry.hub.docker.com/library/busybox:1.25"
	helloworldImage             = "registry.hub.docker.com/library/hello-world"
	helloworldImageLatest       = "registry.hub.docker.com/library/hello-world:latest"
	helloworldImageOnlyRepoName = "hello-world"

	GateWay = "192.168.1.1"
	Subnet  = "192.168.1.0/24"

	testHubAddress = "registry.hub.docker.com"
	testHubUser    = "pouchcontainertest"
	testHubPasswd  = "pouchcontainertest"
)

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
