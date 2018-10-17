package ctrd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_containerLock_TrylockWithRetry(t *testing.T) {
	l := &containerLock{
		ids: make(map[string]struct{}),
	}

	// basically, if the releaseTimeout < the tryLockTimeout,
	// TrylockWithTimeout will lock successfully. If not, it will fail.
	runTrylockWithT := func(tryLockTimeout, releaseTimeout time.Duration) bool {
		id := "c"
		assert.Equal(t, l.Trylock(id), true)
		defer l.Unlock(id)

		var (
			releaseCh = make(chan struct{})
			waitCh    = make(chan bool)
			res       bool
		)

		go func() {
			close(releaseCh)
			ctx, cancel := context.WithTimeout(context.TODO(), tryLockTimeout)
			defer cancel()
			waitCh <- l.TrylockWithRetry(ctx, id)
		}()

		<-releaseCh
		time.Sleep(releaseTimeout)
		l.Unlock(id)

		select {
		case <-time.After(3 * time.Second):
			t.Fatalf("timeout to get the Trylock result")
		case res = <-waitCh:
		}
		return res
	}

	assert.Equal(t, true, runTrylockWithT(5*time.Second, 200*time.Millisecond))
	assert.Equal(t, false, runTrylockWithT(200*time.Millisecond, 500*time.Millisecond))
}

func Test_containerLock_Trylock(t *testing.T) {
	l := &containerLock{
		ids: make(map[string]struct{}),
	}

	assert.Equal(t, len(l.ids), 0)

	// lock a new element
	ok := l.Trylock("element1")
	assert.Equal(t, ok, true)
	assert.Equal(t, len(l.ids), 1)
	assert.Equal(t, l.ids["element1"], struct{}{})

	// lock an existent element
	ok = l.Trylock("element1")
	assert.Equal(t, ok, false)
	assert.Equal(t, len(l.ids), 1)
	assert.Equal(t, l.ids["element1"], struct{}{})

	// lock another new element
	ok = l.Trylock("element2")
	assert.Equal(t, ok, true)
	assert.Equal(t, len(l.ids), 2)
	assert.Equal(t, l.ids["element1"], struct{}{})
}

func Test_containerLock_Unlock(t *testing.T) {
	l := &containerLock{
		ids: make(map[string]struct{}),
	}

	// unlock a non-existent element
	l.Unlock("non-existent")
	assert.Equal(t, len(l.ids), 0)

	// lock a new element
	ok := l.Trylock("element1")
	assert.Equal(t, ok, true)
	assert.Equal(t, len(l.ids), 1)
	assert.Equal(t, l.ids["element1"], struct{}{})

	// unlock an existent element
	l.Unlock("element1")
	assert.Equal(t, len(l.ids), 0)
}
