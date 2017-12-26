package types

import "github.com/docker/libnetwork"

// Network defines the network struct.
type Network struct {
	Name string
	ID   string
	Type string
	Mode string

	Network libnetwork.Network
}
