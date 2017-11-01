package collect

import "sync"

// SafeMap is multiple thread safe, could be accessed by multiple thread simultaneously.
type SafeMap struct {
	sync.RWMutex
	inner map[string]interface{}
}

// NewSafeMap generate a instance of SafeMap type.
func NewSafeMap() *SafeMap {
	return &SafeMap{
		inner: make(map[string]interface{}, 16),
	}
}

// Get return a value from inner map safely.
func (m *SafeMap) Get(k string) *Value {
	m.RLock()
	defer m.RUnlock()

	v, ok := m.inner[k]

	return &Value{v, ok}
}

// Put store a key-value pair into inner map safely.
func (m *SafeMap) Put(k string, v interface{}) {
	m.Lock()
	defer m.Unlock()
	m.inner[k] = v
}

// Remove removes the key-value pair.
func (m *SafeMap) Remove(k string) {
	delete(m.inner, k)
}

// Value represents the value's info of a key-value pair.
type Value struct {
	data interface{}
	ok   bool
}

// Result return the origin data and status in map.
func (v *Value) Result() (interface{}, bool) {
	return v.data, v.ok
}

// Exist return the data exist in map or not.
func (v *Value) Exist() bool {
	return v.ok
}

// String return data as string.
func (v *Value) String() (string, bool) {
	if !v.ok || v.data == nil {
		return "", v.ok
	}
	return v.data.(string), v.ok
}

// Int return data as int.
func (v *Value) Int() (int, bool) {
	if !v.ok || v.data == nil {
		return 0, v.ok
	}
	return v.data.(int), v.ok
}

// Int32 return data as int32.
func (v *Value) Int32() (int32, bool) {
	if !v.ok || v.data == nil {
		return 0, v.ok
	}
	return v.data.(int32), v.ok
}

// Int64 return data as int64.
func (v *Value) Int64() (int64, bool) {
	if !v.ok || v.data == nil {
		return 0, v.ok
	}
	return v.data.(int64), v.ok
}
