package kmutex

import (
	"sync"
	"sync/atomic"
	"time"
)

type value struct {
	c     chan struct{}
	waits int32
}

// KMutex is a lock implement. One key can acquire only one lock. There is no
// lock between different keys.
type KMutex struct {
	sync.Mutex
	keys map[string]*value
}

// New creates a KMutex instance.
func New() *KMutex {
	m := &KMutex{
		keys: make(map[string]*value),
	}

	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C

			m.Mutex.Lock()

			for k, v := range m.keys {
				if v.waits == 0 {
					delete(m.keys, k)
					close(v.c)
				}
			}

			m.Mutex.Unlock()
		}
	}()

	return m
}

func (m *KMutex) lock(k string) (*value, bool) {
	v, ok := m.keys[k]
	if !ok {
		m.keys[k] = &value{
			c:     make(chan struct{}, 1),
			waits: 0,
		}
		return nil, true
	}

	return v, false
}

// Trylock trys to lock, will not block.
func (m *KMutex) Trylock(k string) bool {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	v, ok := m.lock(k)
	if ok {
		return true
	}

	select {
	case <-v.c:
		return true
	default:
		return false
	}
}

// LockWithTimeout trys to lock, if can't acquire the lock, will block util timeout.
func (m *KMutex) LockWithTimeout(k string, to time.Duration) bool {
	m.Mutex.Lock()

	v, ok := m.lock(k)
	if ok {
		m.Mutex.Unlock()
		return true
	}

	atomic.AddInt32(&v.waits, 1)
	defer atomic.AddInt32(&v.waits, -1)

	m.Mutex.Unlock()

	select {
	case <-v.c:
		return true
	case <-time.After(to):
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
func (m *KMutex) Unlock(k string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	v, ok := m.keys[k]
	if ok && len(v.c) == 0 {
		v.c <- struct{}{}
	}
}
