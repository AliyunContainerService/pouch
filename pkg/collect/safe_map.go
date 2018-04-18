package collect

import "sync"

// SafeMap is multiple thread safe, could be accessed by multiple thread simultaneously.
type SafeMap struct {
	sync.RWMutex
	inner map[string]interface{}
}

// NewSafeMap generates a instance of SafeMap type.
func NewSafeMap() *SafeMap {
	return &SafeMap{
		inner: make(map[string]interface{}, 16),
	}
}

// Get returns a value from inner map safely.
func (m *SafeMap) Get(k string) *Value {
	m.RLock()
	defer m.RUnlock()

	v, ok := m.inner[k]

	return &Value{v, ok}
}

// Values returns all key-values stored in map
func (m *SafeMap) Values() map[string]interface{} {
	m.RLock()
	defer m.RUnlock()

	nmap := make(map[string]interface{})
	for k, v := range m.inner {
		nmap[k] = v
	}

	return nmap
}

// Put stores a key-value pair into inner map safely.
func (m *SafeMap) Put(k string, v interface{}) {
	m.Lock()
	defer m.Unlock()

	if m.inner == nil {
		return
	}

	m.inner[k] = v
}

// Remove removes the key-value pair.
func (m *SafeMap) Remove(k string) {
	m.Lock()
	defer m.Unlock()
	delete(m.inner, k)
}

// Value represents the value's info of a key-value pair.
type Value struct {
	data interface{}
	ok   bool
}

// Result returns the origin data and status in map.
func (v *Value) Result() (interface{}, bool) {
	return v.data, v.ok
}

// Exist returns the data exist in map or not.
func (v *Value) Exist() bool {
	return v.ok
}

// String returns data as string.
func (v *Value) String() (string, bool) {
	if !v.ok || v.data == nil {
		return "", v.ok
	}
	if result, ok := v.data.(string); ok {
		return result, ok
	}
	return "", false
}

// Int returns data as int.
func (v *Value) Int() (int, bool) {
	if !v.ok || v.data == nil {
		return 0, v.ok
	}
	if result, ok := v.data.(int); ok {
		return result, ok
	}
	return 0, false
}

// Int32 returns data as int32.
func (v *Value) Int32() (int32, bool) {
	if !v.ok || v.data == nil {
		return 0, v.ok
	}
	if result, ok := v.data.(int32); ok {
		return result, ok
	}
	return 0, false
}

// Int64 returns data as int64.
func (v *Value) Int64() (int64, bool) {
	if !v.ok || v.data == nil {
		return 0, v.ok
	}
	if result, ok := v.data.(int64); ok {
		return result, ok
	}
	return 0, false
}
