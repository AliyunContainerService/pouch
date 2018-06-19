package syslog

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/daemon/logger"

	"github.com/RackSec/srslog"
	pkgerrors "github.com/pkg/errors"
)

var (
	// ErrInvalidSyslogFacility represents the invalid facility.
	ErrInvalidSyslogFacility = errors.New("invalid syslog facility")

	// ErrInvalidSyslogFormat represents the invalid format.
	ErrInvalidSyslogFormat = errors.New("invalid syslog format")

	// ErrFailedToLoadX509KeyPair is to used to indicate that it's failed to load x509 key pair.
	ErrFailedToLoadX509KeyPair = errors.New("fail to load x509 key pair")

	fmtErrInvalidAddressFormat = "syslog-address must be in form proto://address, but got %v"
)

// ValidateSyslogOption validates the syslog config.
func ValidateSyslogOption(info logger.Info) error {
	_, err := parseOptions(info)
	return err
}

// parseFacility parses facility into syslog priority.
func parseFacility(f string) (srslog.Priority, error) {
	if f == "" {
		return defaultSyslogPriority, nil
	}

	if priority, ok := facilityAliasMap[f]; ok {
		return priority, nil
	}

	fInt, err := strconv.Atoi(f)
	if err == nil && 0 <= fInt && fInt <= 23 {
		return srslog.Priority(fInt << 3), nil
	}
	return srslog.Priority(0), ErrInvalidSyslogFacility
}

// parseTargetAddress parses the address into proto and host:port or path.
func parseTargetAddress(address string) (string, string, error) {
	if address == "" {
		return "", "", nil
	}

	if !isTransportURL(address) {
		return "", "", fmt.Errorf(fmtErrInvalidAddressFormat, address)
	}

	url, err := url.Parse(address)
	if err != nil {
		return "", "", err
	}

	switch url.Scheme {
	case "unix", "unixgram":
		if _, err := os.Stat(url.Path); err != nil {
			return "", "", err
		}
		return url.Scheme, url.Path, nil
	default:
		h := url.Host
		if _, _, err := net.SplitHostPort(h); err != nil {
			if !strings.Contains(err.Error(), "missing port in address") {
				return "", "", err
			}
			h = h + ":514"
		}
		return url.Scheme, h, nil
	}
}

// parseLogFormat parse log format into syslog formatter and framer.
func parseLogFormat(logFmt, proto string) (srslog.Formatter, srslog.Framer, error) {
	switch logFmt {
	case "":
		return srslog.UnixFormatter, srslog.DefaultFramer, nil
	case "rfc3164":
		return srslog.RFC3164Formatter, srslog.DefaultFramer, nil
	case "rfc5424":
		if proto == secureProto {
			return rfc5424FormatterWithTagAsAppName, srslog.RFC5425MessageLengthFramer, nil
		}
		return rfc5424FormatterWithTagAsAppName, srslog.DefaultFramer, nil
	case "rfc5424micro":
		if proto == secureProto {
			return rfc5424MicroFormatterWithTagAsAppName, srslog.RFC5425MessageLengthFramer, nil
		}
		return rfc5424MicroFormatterWithTagAsAppName, srslog.DefaultFramer, nil
	default:
		return nil, nil, ErrInvalidSyslogFormat
	}
}

// parseTLSConfig parses the config into tls.Config.
func parseTLSConfig(info logger.Info) (*tls.Config, error) {
	var (
		skipVerify bool
		caFile     string = info.LogConfig["syslog-tls-ca-cert"]
		certFile   string = info.LogConfig["syslog-tls-cert"]
		keyFile    string = info.LogConfig["syslog-tls-key"]
		err        error
	)

	_, skipVerify = info.LogConfig["syslog-tls-skip-verify"]

	tlsCfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		CipherSuites:       defaultCipherSuites,
		InsecureSkipVerify: skipVerify,
	}

	if !skipVerify && caFile != "" {
		pool, err := getCertPool(caFile)
		if err != nil {
			return nil, err
		}
		tlsCfg.RootCAs = pool
	}

	if tlsCfg.Certificates, err = getCert(certFile, keyFile); err != nil {
		return nil, err
	}
	return tlsCfg, nil
}

// isTransportURL returns true if the address has the prefix like
// tcp|udp|unix|unixgram proto.
func isTransportURL(address string) bool {
	for _, pre := range validTransportURLPrefix {
		if strings.HasPrefix(address, pre) {
			return true
		}
	}
	return false
}

// getCertPool returns an X.509 certificate pool from the certificate file.
func getCertPool(caFile string) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to read system certificates: %v", err)
	}

	pem, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate %v: %v", caFile, err)
	}

	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("failed to append certificates from the PEM file: %v", caFile)
	}
	return pool, nil
}

// getCert returns the certificate.
func getCert(certFile string, keyFile string) ([]tls.Certificate, error) {
	if certFile == "" && keyFile == "" {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, pkgerrors.Wrap(err, ErrFailedToLoadX509KeyPair.Error())
	}
	return []tls.Certificate{cert}, nil
}
