package credential

import "github.com/alibaba/pouch/apis/types"

// Save saves a registry credential into a credential store.
func Save(authConfig *types.AuthConfig) error {
	s := loadCredentialStore()
	return s.Save(authConfig)
}

// Get gets a registry credential from a credential store.
func Get(serverAddress string) (types.AuthConfig, error) {
	s := loadCredentialStore()
	return s.Get(serverAddress)
}

// Delete deletes a registry credential from a credential store.
func Delete(serverAddress string) error {
	s := loadCredentialStore()
	return s.Delete(serverAddress)
}

func loadCredentialStore() Store {
	return newFileStore()
}
