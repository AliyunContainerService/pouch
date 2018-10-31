package meta

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/tchap/go-patricia/patricia"
)

var (
	// MetaJSONFile is the default name of json file.
	MetaJSONFile = "meta.json"
)

// Bucket is a Bucket.
type Bucket struct {
	Name string
	Type reflect.Type
}

// NewObject returns Object.
func (b *Bucket) NewObject() Object {
	return reflect.New(b.Type).Interface().(Object)
}

// Config represents the configurations used for metadata store.
type Config struct {
	Driver  string
	Buckets []Bucket
	BaseDir string
}

// Object is an interface.
type Object interface {
	Key() string
}

// Store defines what a metadata store should be like.
type Store struct {
	Config
	trieLock *sync.Mutex // trieLock use to protect 'trie'.
	trie     *patricia.Trie
	current  *Bucket
	backend  Backend
}

// NewStore creates a backend storage.
func NewStore(cfg Config) (*Store, error) {
	if len(cfg.Buckets) < 1 {
		return nil, fmt.Errorf("config bucket can not be empty")
	}

	if cfg.Driver == "" {
		cfg.Driver = DefaultStore
	}

	if _, ok := backendFactory[cfg.Driver]; !ok {
		return nil, fmt.Errorf("store driver %s not found", cfg.Driver)
	}
	backend, err := backendFactory[cfg.Driver](cfg)
	if err != nil {
		return nil, fmt.Errorf("create driver %s failed: %v", cfg.Driver, err)
	}

	s := &Store{
		Config:   cfg,
		backend:  backend,
		trieLock: new(sync.Mutex),
		trie:     patricia.NewTrie(),
	}

	keys := []string{}
	for _, bucket := range cfg.Buckets {
		k, err := s.backend.Keys(bucket.Name)
		if err != nil {
			return nil, err
		}

		keys = append(keys, k...)
	}

	for _, key := range keys {
		s.trie.Insert(patricia.Prefix(key), struct{}{})
	}

	s = s.Bucket(cfg.Buckets[0].Name)
	if s == nil {
		return nil, fmt.Errorf("failed to new bucket store")
	}

	return s, nil
}

// Bucket returns a specific store instance.
// And name is used to specify the bucket's name that will be write.
func (s *Store) Bucket(name string) *Store {
	if name == "" {
		if s.current == nil {
			return nil
		}
		return s
	}

	var pb *Bucket

	for _, b := range s.Buckets {
		if name == b.Name {
			pb = &b
			break
		}
	}
	if pb == nil {
		return nil
	}

	return &Store{
		Config:   s.Config,
		backend:  s.backend,
		current:  pb,
		trieLock: s.trieLock,
		trie:     s.trie,
	}
}

// Put writes the 'obj' into backend storage.
func (s *Store) Put(obj Object) error {
	value, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to encode meta data: %v", err)
	}
	if err := s.backend.Put(s.current.Name, obj.Key(), value); err != nil {
		return fmt.Errorf("failed to put meta data: %v", err)
	}

	// add key into trie tree.
	s.trieLock.Lock()
	s.trie.Insert(patricia.Prefix(obj.Key()), struct{}{})
	s.trieLock.Unlock()

	return nil
}

// Fetch uses to get meta data and decode it into 'obj'.
func (s *Store) Fetch(obj Object) error {
	value, err := s.backend.Get(s.current.Name, obj.Key())
	if err != nil {
		return err
	}

	if err := json.Unmarshal(value, obj); err != nil {
		return fmt.Errorf("failed to decode meta data: %v", err)
	}
	return nil
}

// Get uses to get meta data.
func (s *Store) Get(key string) (Object, error) {
	value, err := s.backend.Get(s.current.Name, key)
	if err != nil {
		return nil, err
	}

	obj := s.current.NewObject()
	if err := json.Unmarshal(value, obj); err != nil {
		return nil, fmt.Errorf("failed to decode meta data: %v", err)
	}
	return obj, nil
}

// Remove calls it to remove a object.
func (s *Store) Remove(key string) error {
	if err := s.backend.Remove(s.current.Name, key); err != nil {
		return err
	}

	// delete key from trie tree.
	s.trieLock.Lock()
	s.trie.Delete(patricia.Prefix(key))
	s.trieLock.Unlock()

	return nil
}

// ForEach uses to handle every object by sequence.
func (s *Store) ForEach(handle func(Object) error) error {
	values, err := s.backend.List(s.current.Name)
	if err != nil {
		return fmt.Errorf("failed to list meta data: %v", err)
	}
	for _, v := range values {
		obj := s.current.NewObject()
		if err := json.Unmarshal(v, obj); err != nil {
			return fmt.Errorf("failed to decode meta data: %v", err)
		}
		if err := handle(obj); err != nil {
			return err
		}
	}
	return nil
}

// List returns all objects as map.
func (s *Store) List() (map[string]Object, error) {
	objs := make(map[string]Object, 32)

	handle := func(obj Object) error {
		objs[obj.Key()] = obj
		return nil
	}
	if err := s.ForEach(handle); err != nil {
		return nil, err
	}
	return objs, nil
}

// Keys returns all keys only.
func (s *Store) Keys() ([]string, error) {
	keys, err := s.backend.Keys(s.current.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %v", err)
	}
	return keys, nil
}

// GetWithPrefix return objects matching prefix.
func (s *Store) GetWithPrefix(prefix string) ([]Object, error) {
	keys, err := s.KeysWithPrefix(prefix)
	if err != nil {
		return nil, err
	}

	var objects []Object

	for _, key := range keys {
		obj, err := s.Get(key)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

// KeysWithPrefix return all keys matching prefix.
func (s *Store) KeysWithPrefix(prefix string) ([]string, error) {
	var keys []string

	if len(prefix) == 0 {
		return keys, nil
	}

	fn := func(prefix patricia.Prefix, item patricia.Item) error {
		keys = append(keys, string(prefix))
		return nil
	}

	s.trieLock.Lock()
	defer s.trieLock.Unlock()

	if err := s.trie.VisitSubtree(patricia.Prefix(prefix), fn); err != nil {
		return keys, err
	}
	return keys, nil
}

// Path returns the path with specified key.
func (s *Store) Path(key string) string {
	return s.backend.Path(key)
}

// Shutdown releases all resources used by the backend
func (s *Store) Shutdown() error {
	return s.backend.Close()
}
