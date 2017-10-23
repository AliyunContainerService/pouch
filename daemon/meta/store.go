package meta

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	current *Bucket
	backend Backend
}

// NewStore creates a backend storage.
func NewStore(cfg Config) (*Store, error) {
	s := &Store{
		Config:  cfg,
		backend: backend,
	}

	if err := s.backend.New(cfg); err != nil {
		return nil, err
	}

	return s.Bucket(MetaJSONFile), nil
}

// Bucket returns a specific store instance.
// And name is used to specify the file'name that will be write.
func (s *Store) Bucket(name string) *Store {
	if name == "" {
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
		Config:  s.Config,
		backend: s.backend,
		current: pb,
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
	return nil
}

// Fetch uses to get meta data and decode it into 'obj'.
func (s *Store) Fetch(obj Object) error {
	value, err := s.backend.Get(s.current.Name, obj.Key())
	if err != nil {
		return fmt.Errorf("failed to get meta data: %v", err)
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
		return nil, fmt.Errorf("failed to get meta data: %v", err)
	}

	obj := s.current.NewObject()
	if err := json.Unmarshal(value, obj); err != nil {
		return nil, fmt.Errorf("failed to decode meta data: %v", err)
	}
	return obj, nil
}

// Remove calls it to remove a object.
func (s *Store) Remove(key string) error {
	return s.backend.Remove(key)
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
	keys, err := s.backend.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %v", err)
	}
	return keys, nil
}
