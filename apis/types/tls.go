package types

// TLSConfig contains information of tls which users can specify
type TLSConfig struct {
	CA           string
	Cert         string
	Key          string
	VerifyRemote bool
}
