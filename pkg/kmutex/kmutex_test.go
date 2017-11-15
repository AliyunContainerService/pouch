package kmutex

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestKMutex(t *testing.T) {
	m := New()

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			k := strconv.Itoa(i)

			if m.Trylock(k) == false {
				t.Fatalf("failed to trylock: %d", i)
			}
			if m.Trylock(k) == true {
				t.Fatalf("trylock is error: %d", i)
			}

			m.Unlock(k)

			if m.Trylock(k) == false {
				t.Fatalf("failed to trylock: %d", i)
			}
		}(i)
	}

	wg.Wait()
}

func TestKMutexTimeout(t *testing.T) {
	m := New()

	running := make(chan struct{})
	wait := make(chan struct{})
	go func() {
		if m.Trylock("key") == false {
			t.Fatalf("failed to trylock")
		}

		close(running)

		time.Sleep(time.Second * 10)
		close(wait)
	}()

	<-running

	go func() {
		<-wait
		t.Fatalf("failed to trylock with timeout")
	}()

	if m.LockWithTimeout("key", time.Second*5) == true {
		t.Fatalf("trylock with timeout is error")
	}
}
