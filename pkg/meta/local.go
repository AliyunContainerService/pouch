package meta

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

func init() {
	Register("local", NewLocalStore)
}

type localStore struct {
	sync.Mutex
	base  string
	cache map[string][]byte
}

// NewLocalStore is used to make local metadata store instance.
func NewLocalStore(cfg Config) (Backend, error) {
	if !path.IsAbs(cfg.BaseDir) {
		return nil, fmt.Errorf("Not absolute path: %s", cfg.BaseDir)
	}
	if err := mkdirIfNotExist(cfg.BaseDir); err != nil {
		return nil, err
	}

	s := &localStore{
		cache: make(map[string][]byte, 64),
		base:  cfg.BaseDir,
	}

	// initialize cache
	handle := func(f os.FileInfo) error {
		if _, err := s.Get(MetaJSONFile, f.Name()); err != nil {
			return err
		}
		// TODO maybe get other file.

		return nil
	}

	if err := walkDir(s.base, handle); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *localStore) Path(key string) string {
	return path.Join(s.base, key)
}

func (s *localStore) Put(fileName, key string, value []byte) error {
	dir := path.Join(s.base, key)

	s.Lock()
	defer s.Unlock()

	if err := mkdirIfNotExist(dir); err != nil {
		return err
	}

	name := filepath.Join(dir, fileName)
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %s, %v", name, err)
	}
	defer f.Close()

	if _, err := f.Write(value); err != nil {
		return fmt.Errorf("failed to write file: %s, %v", name, err)
	}
	f.Sync()

	// NOTICE: cache the key-value.
	s.cache[key+"/"+fileName] = value

	return nil
}

func (s *localStore) Get(fileName, key string) ([]byte, error) {
	s.Lock()
	defer s.Unlock()

	// NOTICE: find cache firstly.
	v, ok := s.cache[key+"/"+fileName]
	if ok {
		return v, nil
	}

	dir := path.Join(s.base, key)
	name := filepath.Join(dir, fileName)

	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrObjectNotFound
		}
	}

	value, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s, %v", name, err)
	}

	// NOTICE: cache the key-value.
	s.cache[key+"/"+fileName] = value

	return value, nil
}

func (s *localStore) Remove(bucket string, key string) error {
	dir := path.Join(s.base, key)

	s.Lock()
	defer s.Unlock()

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove dir: %s, %v", dir, err)
	}

	// NOTICE: delete cache
	for k := range s.cache {
		if strings.HasPrefix(k, key+"/") {
			delete(s.cache, k)
		}
	}

	return nil
}

func (s *localStore) List(fileName string) ([][]byte, error) {
	s.Lock()
	defer s.Unlock()

	values := make([][]byte, 0, len(s.cache))

	for k, v := range s.cache {
		if strings.HasSuffix(k, "/"+fileName) {
			values = append(values, v)
		}
	}

	return values, nil
}

func (s *localStore) Keys(fileName string) ([]string, error) {
	s.Lock()
	defer s.Unlock()

	keys := make([]string, 0, len(s.cache))

	for k := range s.cache {
		fields := strings.Split(k, "/")
		if len(fields) != 2 {
			return nil, fmt.Errorf("failed to split cache key: %s", k)
		}
		keys = append(keys, fields[0])
	}

	return keys, nil
}

func mkdirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0744); err != nil {
				return fmt.Errorf("failed to mkdir %s: %v", dir, err)
			}
		}
	}
	return nil
}

func walkDir(dir string, handle func(os.FileInfo) error) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read dir %s: %v", dir, err)
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		if err := handle(f); err != nil {
			return fmt.Errorf("failed to handle file %s: %v", f.Name(), err)
		}
	}
	return nil
}
