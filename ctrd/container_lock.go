package ctrd

import (
	"sync"
)

// containerLock use to make sure that only one operates the container at the same time.
type containerLock struct {
	mutex sync.Mutex
	ids   map[string]struct{}
}

func (l *containerLock) Trylock(id string) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	_, ok := l.ids[id]
	if !ok {
		l.ids[id] = struct{}{}
	}
	return !ok
}

func (l *containerLock) Unlock(id string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	delete(l.ids, id)
}
