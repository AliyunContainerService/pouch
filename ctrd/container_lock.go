package ctrd

import (
	"context"
	"math/rand"
	"sync"
	"time"
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
		return true
	}
	return false
}

func (l *containerLock) Unlock(id string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	delete(l.ids, id)
}

func (l *containerLock) TrylockWithRetry(ctx context.Context, id string) bool {
	var retry = 32

	for {
		ok := l.Trylock(id)
		if ok {
			return true
		}

		// sleep random duration by retry
		select {
		case <-time.After(time.Millisecond * time.Duration(rand.Intn(retry))):
			if retry < 2048 {
				retry = retry << 1
			}
			continue
		case <-ctx.Done():
			return false
		}
	}
}
