package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/spf13/pflag"
)

// TLSConfig contains information of tls which users can specify
type TLSConfig struct {
	CA           string
	Cert         string
	Key          string
	VerifyRemote bool
}

// GenTLSConfig returns a tls config object according to inputting parameters.
func GenTLSConfig(key, cert, ca string) (*tls.Config, error) {
	tlsConfig := &tls.Config{}
	tlsCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to read X509 key pair (cert: %q, key: %q): %v", cert, key, err)
	}
	tlsConfig.Certificates = []tls.Certificate{tlsCert}
	if ca != "" {
		cp, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to read system certificates: %v", err)
		}
		pem, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate %q: %v", ca, err)
		}
		if !cp.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("failed to append certificates from PEM file: %q", ca)
		}

		tlsConfig.ClientCAs = cp
	}

	return tlsConfig, nil
}

// SetupTLSFlag setups flags of tls arguments
func SetupTLSFlag(fs *pflag.FlagSet, tlsCfg *TLSConfig) {
	fs.StringVar(&tlsCfg.Key, "tlskey", "", "Specify key file of tls")
	fs.StringVar(&tlsCfg.Cert, "tlscert", "", "Specify cert file of tls")
	fs.StringVar(&tlsCfg.CA, "tlscacert", "", "Specify CA file of tls")
	fs.BoolVar(&tlsCfg.VerifyRemote, "tlsverify", false, "Switch if verify the remote when using tls")
}
