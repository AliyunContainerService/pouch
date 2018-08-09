package syslog

import (
	"crypto/tls"
	"errors"
	"net"
)

type serverConn interface {
	writeString(framer Framer, formatter Formatter, p Priority, hostname, tag, s string) error
	close() error
}

func makeDialer(proto string, addr string, cfg *tls.Config) (serverConn, string, error) {
	switch proto {
	case "":
		return unixLocalDialer()
	case secureProto:
		return tlsDialer(addr, cfg)
	default:
		return commonDialer(proto, addr)
	}
}

// commonDialer is the most common dialer for TCP/UDP/Unix connections.
func commonDialer(network string, addr string) (serverConn, string, error) {
	var (
		sc       serverConn
		hostname string
	)

	c, err := net.Dial(network, addr)
	if err == nil {
		sc = &remoteConn{conn: c}
		hostname = c.LocalAddr().String()
	}
	return sc, hostname, err
}

// tlsDialer connects to TLS over TCP, and is used for the "tcp+tls" network.
func tlsDialer(addr string, cfg *tls.Config) (serverConn, string, error) {
	var (
		sc       serverConn
		hostname string
	)

	c, err := tls.Dial("tcp", addr, cfg)
	if err == nil {
		sc = &remoteConn{conn: c}
		hostname = c.LocalAddr().String()
	}
	return sc, hostname, err
}

// unixLocalDialer opens a Unix domain socket connection to the syslog daemon
// running on the local machine.
func unixLocalDialer() (serverConn, string, error) {
	for _, network := range unixDialerTypes {
		for _, path := range unixDialerLocalPaths {
			conn, err := net.Dial(network, path)
			if err == nil {
				return &localConn{conn: conn}, "localhost", nil
			}
		}
	}
	return nil, "", errors.New("unix local syslog delivery error")
}
