package opts

import (
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"
)

// ValidateExtraHost validates the provided string is a valid extra-host.
func ValidateExtraHost(val string) error {
	// allow for IPv6 addresses in extra hosts by only splitting on first ":"
	arr := strings.SplitN(val, ":", 2)
	if len(arr) != 2 || len(arr[0]) == 0 {
		return errors.Errorf("bad format for add-host: %q", val)
	}
	// TODO(lang710): Skip ipaddr validation for special "host-gateway" string
	//  If the IP Address is a string called "host-gateway", replace this
	//  value with the IP address stored in the daemon level HostGatewayIP
	//  config variable
	if _, err := validateIPAddress(arr[1]); err != nil {
		return errors.Wrapf(err, "invalid IP address in add-host: %q", arr[1])
	}
	return nil
}

// validateIPAddress validates an Ip address.
func validateIPAddress(val string) (string, error) {
	var ip = net.ParseIP(strings.TrimSpace(val))
	if ip != nil {
		return ip.String(), nil
	}
	return "", fmt.Errorf("%s is not an ip address", val)
}
