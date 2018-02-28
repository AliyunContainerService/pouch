package credential

import "github.com/alibaba/pouch/apis/types"

// Store implements storage type of registry credentials.
type Store interface {
	// Save saves a credential into a Store.
	Save(authConfig *types.AuthConfig) error

	// Get gets a credential from a Store.
	Get(serverAddress string) (types.AuthConfig, error)

	// Delete deletes a credential in Store.
	Delete(serverAddress string) error

	// Exist determines whether a specified credential is exist in Store.
	Exist(serverAddress string) bool
}
