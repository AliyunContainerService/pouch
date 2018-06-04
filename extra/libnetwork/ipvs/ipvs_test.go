// +build linux

package ipvs

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"github.com/docker/libnetwork/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
)

var (
	schedMethods = []string{
		RoundRobin,
		LeastConnection,
		DestinationHashing,
		SourceHashing,
	}

	protocols = []string{
		"TCP",
		"UDP",
		"FWM",
	}

	fwdMethods = []uint32{
		ConnectionFlagMasq,
		ConnectionFlagTunnel,
		ConnectionFlagDirectRoute,
	}

	fwdMethodStrings = []string{
		"Masq",
		"Tunnel",
		"Route",
	}
)

func checkDestination(t *testing.T, checkPresent bool, protocol, serviceAddress, realAddress, fwdMethod string) {
	var (
		realServerStart bool
		realServers     []string
	)

	out, err := exec.Command("ipvsadm", "-Ln").CombinedOutput()
	require.NoError(t, err)

	for _, o := range strings.Split(string(out), "\n") {
		cmpStr := serviceAddress
		if protocol == "FWM" {
			cmpStr = " " + cmpStr
		}

		if strings.Contains(o, cmpStr) {
			realServerStart = true
			continue
		}

		if realServerStart {
			if !strings.Contains(o, "->") {
				break
			}

			realServers = append(realServers, o)
		}
	}

	for _, r := range realServers {
		if strings.Contains(r, realAddress) {
			parts := strings.Fields(r)
			assert.Equal(t, fwdMethod, parts[2])
			return
		}
	}

	if checkPresent {
		t.Fatalf("Did not find the destination %s fwdMethod %s in ipvs output", realAddress, fwdMethod)
	}
}

func checkService(t *testing.T, checkPresent bool, protocol, schedMethod, serviceAddress string) {
	out, err := exec.Command("ipvsadm", "-Ln").CombinedOutput()
	require.NoError(t, err)

	for _, o := range strings.Split(string(out), "\n") {
		cmpStr := serviceAddress
		if protocol == "FWM" {
			cmpStr = " " + cmpStr
		}

		if strings.Contains(o, cmpStr) {
			parts := strings.Split(o, " ")
			assert.Equal(t, protocol, parts[0])
			assert.Equal(t, serviceAddress, parts[2])
			assert.Equal(t, schedMethod, parts[3])

			if !checkPresent {
				t.Fatalf("Did not expect the service %s in ipvs output", serviceAddress)
			}

			return
		}
	}

	if checkPresent {
		t.Fatalf("Did not find the service %s in ipvs output", serviceAddress)
	}
}

func TestGetFamily(t *testing.T) {
	if testutils.RunningOnCircleCI() {
		t.Skipf("Skipping as not supported on CIRCLE CI kernel")
	}

	id, err := getIPVSFamily()
	require.NoError(t, err)
	assert.NotEqual(t, 0, id)
}

func TestService(t *testing.T) {
	if testutils.RunningOnCircleCI() {
		t.Skipf("Skipping as not supported on CIRCLE CI kernel")
	}

	defer testutils.SetupTestOSContext(t)()

	i, err := New("")
	require.NoError(t, err)

	for _, protocol := range protocols {
		for _, schedMethod := range schedMethods {
			var serviceAddress string

			s := Service{
				AddressFamily: nl.FAMILY_V4,
				SchedName:     schedMethod,
			}

			switch protocol {
			case "FWM":
				s.FWMark = 1234
				serviceAddress = fmt.Sprintf("%d", 1234)
			case "TCP":
				s.Protocol = syscall.IPPROTO_TCP
				s.Port = 80
				s.Address = net.ParseIP("1.2.3.4")
				s.Netmask = 0xFFFFFFFF
				serviceAddress = "1.2.3.4:80"
			case "UDP":
				s.Protocol = syscall.IPPROTO_UDP
				s.Port = 53
				s.Address = net.ParseIP("2.3.4.5")
				serviceAddress = "2.3.4.5:53"
			}

			err := i.NewService(&s)
			assert.NoError(t, err)
			checkService(t, true, protocol, schedMethod, serviceAddress)
			var lastMethod string
			for _, updateSchedMethod := range schedMethods {
				if updateSchedMethod == schedMethod {
					continue
				}

				s.SchedName = updateSchedMethod
				err = i.UpdateService(&s)
				assert.NoError(t, err)
				checkService(t, true, protocol, updateSchedMethod, serviceAddress)
				lastMethod = updateSchedMethod
			}

			err = i.DelService(&s)
			checkService(t, false, protocol, lastMethod, serviceAddress)
		}
	}

}

func createDummyInterface(t *testing.T) {
	if testutils.RunningOnCircleCI() {
		t.Skipf("Skipping as not supported on CIRCLE CI kernel")
	}

	dummy := &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{
			Name: "dummy",
		},
	}

	err := netlink.LinkAdd(dummy)
	require.NoError(t, err)

	dummyLink, err := netlink.LinkByName("dummy")
	require.NoError(t, err)

	ip, ipNet, err := net.ParseCIDR("10.1.1.1/24")
	require.NoError(t, err)

	ipNet.IP = ip

	ipAddr := &netlink.Addr{IPNet: ipNet, Label: ""}
	err = netlink.AddrAdd(dummyLink, ipAddr)
	require.NoError(t, err)
}

func TestDestination(t *testing.T) {
	defer testutils.SetupTestOSContext(t)()

	createDummyInterface(t)
	i, err := New("")
	require.NoError(t, err)

	for _, protocol := range protocols {
		var serviceAddress string

		s := Service{
			AddressFamily: nl.FAMILY_V4,
			SchedName:     RoundRobin,
		}

		switch protocol {
		case "FWM":
			s.FWMark = 1234
			serviceAddress = fmt.Sprintf("%d", 1234)
		case "TCP":
			s.Protocol = syscall.IPPROTO_TCP
			s.Port = 80
			s.Address = net.ParseIP("1.2.3.4")
			s.Netmask = 0xFFFFFFFF
			serviceAddress = "1.2.3.4:80"
		case "UDP":
			s.Protocol = syscall.IPPROTO_UDP
			s.Port = 53
			s.Address = net.ParseIP("2.3.4.5")
			serviceAddress = "2.3.4.5:53"
		}

		err := i.NewService(&s)
		assert.NoError(t, err)
		checkService(t, true, protocol, RoundRobin, serviceAddress)

		s.SchedName = ""
		for j, fwdMethod := range fwdMethods {
			d1 := Destination{
				AddressFamily:   nl.FAMILY_V4,
				Address:         net.ParseIP("10.1.1.2"),
				Port:            5000,
				Weight:          1,
				ConnectionFlags: fwdMethod,
			}

			realAddress := "10.1.1.2:5000"
			err := i.NewDestination(&s, &d1)
			assert.NoError(t, err)
			checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[j])
			d2 := Destination{
				AddressFamily:   nl.FAMILY_V4,
				Address:         net.ParseIP("10.1.1.3"),
				Port:            5000,
				Weight:          1,
				ConnectionFlags: fwdMethod,
			}

			realAddress = "10.1.1.3:5000"
			err = i.NewDestination(&s, &d2)
			assert.NoError(t, err)
			checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[j])

			d3 := Destination{
				AddressFamily:   nl.FAMILY_V4,
				Address:         net.ParseIP("10.1.1.4"),
				Port:            5000,
				Weight:          1,
				ConnectionFlags: fwdMethod,
			}

			realAddress = "10.1.1.4:5000"
			err = i.NewDestination(&s, &d3)
			assert.NoError(t, err)
			checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[j])

			for m, updateFwdMethod := range fwdMethods {
				if updateFwdMethod == fwdMethod {
					continue
				}
				d1.ConnectionFlags = updateFwdMethod
				realAddress = "10.1.1.2:5000"
				err = i.UpdateDestination(&s, &d1)
				assert.NoError(t, err)
				checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[m])

				d2.ConnectionFlags = updateFwdMethod
				realAddress = "10.1.1.3:5000"
				err = i.UpdateDestination(&s, &d2)
				assert.NoError(t, err)
				checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[m])

				d3.ConnectionFlags = updateFwdMethod
				realAddress = "10.1.1.4:5000"
				err = i.UpdateDestination(&s, &d3)
				assert.NoError(t, err)
				checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[m])
			}

			err = i.DelDestination(&s, &d1)
			assert.NoError(t, err)
			err = i.DelDestination(&s, &d2)
			assert.NoError(t, err)
			err = i.DelDestination(&s, &d3)
			assert.NoError(t, err)
		}
	}
}
