package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/pouch/test/command"
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
	busyboxImage125   = "registry.hub.docker.com/library/busybox:1.25"
	busyboxImage125ID = "sha256:e02e811dd08fd49e7f6032625495118e63f597eb150403d02e3238af1df240ba"
	testHubAddress    = "registry.hub.docker.com"
	testHubUser       = "pouchcontainertest"
	testHubPasswd     = "pouchcontainertest"

	testDaemonHTTPSAddr = "tcp://0.0.0.0:2000"
	serverCa            = "/tmp/tls/server/ca.pem"
	serverCert          = "/tmp/tls/server/cert.pem"
	serverKey           = "/tmp/tls/server/key.pem"
	clientCa            = "/tmp/tls/a_client/ca.pem"
	clientCert          = "/tmp/tls/a_client/cert.pem"
	clientKey           = "/tmp/tls/a_client/key.pem"
	clientWrongCa       = "/tmp/tls/a_client/ca_wrong.pem"
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

type testingTB interface {
	Fatalf(format string, args ...interface{})
	Skip(string)
}

func helpwantedForMissingCase(t testingTB, name string) {
	t.Skip(fmt.Sprintf("help wanted: %s", name))
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

// IsTLSExist check if the TLS related file exists.
func IsTLSExist() bool {
	if _, err := os.Stat(serverCa); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(serverKey); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(serverCert); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(clientCa); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(clientCert); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(clientKey); os.IsNotExist(err) {
		return false
	}
	return true
}

// inspectFilter get the string of info via inspect -f
func inspectFilter(name, filter string) (string, error) {
	format := fmt.Sprintf("{{%s}}", filter)
	result := command.PouchRun("inspect", "-f", format, name)
	if result.Error != nil || result.ExitCode != 0 {
		return "", fmt.Errorf("failed to inspect container %s via filter %s: %s", name, filter, result.Combined())
	}
	return strings.TrimSpace(result.Combined()), nil
}
