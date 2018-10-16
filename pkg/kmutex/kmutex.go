package kmutex

import (
	"sync"
	"sync/atomic"
	"time"
)

// value is the lock details of a race instance.
type value struct {
	// c is channel to identify whether this race instance is released.
	// If c is filled up, it means released.
	c chan struct{}

	// waits means how many callers are waiting for the lock.
	// When caller wants to Lock and has not finished, waits would increase.
	// And when caller finishes to lock (no matter fails or succeeds), it decreases.
	waits int32
}

// KMutex is a lock implement. One key can acquire only one lock.
// There is no lock between different keys.
type KMutex struct {
	sync.Mutex
	// lockedKeys contains the race instances.
	// Assuming that there is a race instance and it is stored in lockedKeys with a key,
	// it means that this race instance is used by one caller, no one can get this instance any longer.
	// If there is no key in lockedKeys representing the race instance, one could get the instance
	// and set the key of it in lockedKeys.
	lockedKeys map[string]*value
}

// New creates a KMutex instance.
func New() *KMutex {
	m := &KMutex{
		lockedKeys: make(map[string]*value),
	}

	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C

			m.Mutex.Lock()
			for k, v := range m.lockedKeys {
				if v.waits == 0 {
					delete(m.lockedKeys, k)
					close(v.c)
				}
			}
			m.Mutex.Unlock()
		}
	}()

	return m
}

// lock adds the key in the lockedKeys.
// It means that key which represents a race instance is accessed by the caller.
func (m *KMutex) lock(key string) (*value, bool) {
	v, ok := m.lockedKeys[key]
	if ok {
		// return false the the value if it has been locked.
		return v, false
	}
	m.lockedKeys[key] = &value{
		c:     make(chan struct{}, 1),
		waits: 0,
	}
	return nil, true
}

// Trylock tries to lock. It will not block the caller.
// No matter lock fails or succeeds, function returns immediately.
func (m *KMutex) Trylock(k string) bool {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	v, ok := m.lock(k)
	if ok {
		return true
	}

	// the k has already been locked.
	select {
	case <-v.c:
		// the locker has released the lock.
		return true
	default:
		// return false immediately if someone locked it.
		return false
	}
}

// LockWithTimeout tries to lock.
// It can't acquire the lock, will block util timeout.
func (m *KMutex) LockWithTimeout(k string, to time.Duration) bool {
	m.Mutex.Lock()

	v, ok := m.lock(k)
	if ok {
		m.Mutex.Unlock()
		return true
	}

	// the k has already been locked.
	atomic.AddInt32(&v.waits, 1)
	defer atomic.AddInt32(&v.waits, -1)

	m.Mutex.Unlock()

	select {
	case <-v.c:
		// the locker has released the lock.
		return true
	case <-time.After(to):
		// timeout before get the released lock.
		return false
	}
}

// Lock waits to acquire lock.
func (m *KMutex) Lock(k string) bool {
	m.Mutex.Lock()

	v, ok := m.lock(k)
	if ok {
		m.Mutex.Unlock()
		return true
	}

	atomic.AddInt32(&v.waits, 1)
	defer atomic.AddInt32(&v.waits, -1)

	m.Mutex.Unlock()

	<-v.c
	return true
}

// Unlock release the lock.
// If the key does not exist, return immediately.
// Otherwise, fill up the chan to broadcast that someone released the lock.
func (m *KMutex) Unlock(k string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	v, ok := m.lockedKeys[k]
	if ok && len(v.c) == 0 {
		// filled up the chan to identify caller has released the lock.
		v.c <- struct{}{}
	}
}
