package credential

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/alibaba/pouch/apis/types"
)

type fileStore struct {
	configFile *ConfigFile
	fileName   string
}

func newFileStore() Store {
	fs := &fileStore{
		fileName: filepath.Join(homedir(), configFileName),
	}

	fd, err := os.Open(fs.fileName)
	if err != nil {
		return fs
	}
	defer fd.Close()

	var configFile ConfigFile
	err = json.NewDecoder(fd).Decode(&configFile)
	if err == nil {
		fs.configFile = &configFile
	}

	return fs
}

// Save implements Store interface.
func (fs *fileStore) Save(authConfig *types.AuthConfig) error {
	if fs.configFile == nil {
		fs.configFile = &ConfigFile{
			AuthConfigs: make(map[string]types.AuthConfig),
		}
	}

	encodedAuth := encodeAuth(authConfig.Username, authConfig.Password)
	if encodedAuth == "" {
		return nil
	}

	serverAddress := authConfig.ServerAddress
	if serverAddress == "" {
		serverAddress = defaultRegistry
	} else {
		serverAddress = addrTrim(serverAddress)
	}

	fs.configFile.AuthConfigs[serverAddress] = types.AuthConfig{
		Auth: encodedAuth,
	}

	return fs.update()
}

// Get implements Store interface.
func (fs *fileStore) Get(serverAddress string) (types.AuthConfig, error) {
	if fs.configFile == nil {
		return types.AuthConfig{}, nil
	}

	authConfigs := fs.configFile.AuthConfigs
	if serverAddress == "" {
		serverAddress = defaultRegistry
	} else {
		serverAddress = addrTrim(serverAddress)
	}
	authConfig, exist := authConfigs[serverAddress]
	if !exist {
		return types.AuthConfig{}, nil
	}

	username, password, err := decodeAuth(authConfig.Auth)
	if err != nil {
		return types.AuthConfig{}, err
	}

	authConfig.Username = username
	authConfig.Password = password
	authConfig.ServerAddress = serverAddress
	return authConfig, nil
}

// Delete implements Store interface.
func (fs *fileStore) Delete(serverAddress string) error {
	if fs.configFile == nil {
		return nil
	}

	serverAddress = addrTrim(serverAddress)
	delete(fs.configFile.AuthConfigs, serverAddress)
	return fs.update()
}

// update updates file store with new contents.
func (fs *fileStore) update() error {
	if fs.configFile == nil {
		return nil
	}

	dir := filepath.Dir(fs.fileName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	fd, err := os.OpenFile(fs.fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer fd.Close()

	data, err := json.MarshalIndent(fs.configFile, "", "    ")
	if err != nil {
		return err
	}

	_, err = fd.Write(data)
	return err
}
