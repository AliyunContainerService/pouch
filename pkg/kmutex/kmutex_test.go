package kmutex

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestKMutex(t *testing.T) {
	m := New()

	var wg sync.WaitGroup
	errchan := make(chan error)

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			k := strconv.Itoa(i)

			if !m.Trylock(k) {
				errchan <- errors.Errorf("failed to trylock: %d", i)
			}
			if m.Trylock(k) {
				errchan <- errors.Errorf("trylock is error: %d", i)
			}

			m.Unlock(k)

			if !m.Trylock(k) {
				errchan <- errors.Errorf("failed to trylock: %d", i)
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(errchan)
	}()

	if err := <-errchan; err != nil {
		t.Fatal(err)
	}
}

func TestKMutexTimeout(t *testing.T) {
	m := New()

	running := make(chan struct{})
	errchan := make(chan error)

	go func() {
		if !m.Trylock("key") {
			errchan <- errors.New("failed to trylock")
			return
		}

		close(running)

		time.Sleep(time.Second * 10)
	}()

	go func() {
		<-running
		close(errchan)
	}()

	if err := <-errchan; err != nil {
		t.Fatal(err)
	}

	if m.LockWithTimeout("key", time.Second*5) {
		t.Fatalf("trylock with timeout is error")
	}
}
