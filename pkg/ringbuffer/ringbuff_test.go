package ringbuffer

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestPushNormal(t *testing.T) {
	count := 5
	rb := New(count)

	// make the buffer full
	for i := 0; i < count; i++ {
		covered, err := rb.Push(i)
		assertHelper(t, false, covered, "unexpected to drop data")
		assertHelper(t, nil, err, "unexpected error during push non-closed queue: %v", err)
	}

	// continue to push new data
	for i := 0; i < count; i++ {
		covered, err := rb.Push(i + count)
		assertHelper(t, true, covered, "expected to drop data, but not")
		assertHelper(t, nil, err, "unexpected error during push non-closed queue: %v", err)
	}

	// check the buffer data
	expectedDump := make([]interface{}, 0, count)
	for i := 0; i < count; i++ {
		expectedDump = append(expectedDump, count+i)
	}

	got := rb.Drain()
	assertHelper(t, expectedDump, got, "expected return %v, but got %v", expectedDump, got)
}

func TestPopNormal(t *testing.T) {
	count := 5
	rb := New(count)

	// make the buffer full
	for i := 0; i < count; i++ {
		covered, err := rb.Push(i)
		assertHelper(t, false, covered, "unexpected to drop data")
		assertHelper(t, nil, err, "unexpected error during push non-closed queue: %v", err)
	}

	for i := 0; i < count; i++ {
		val, err := rb.Pop()
		assertHelper(t, nil, err, "unexpected error during pop: %v", err)
		assertHelper(t, i, val, "expected to have %v, but got %v", i, val)
	}

	assertHelper(t, 0, rb.q.size(), "expected to have empty queue, but got %d size of queue", rb.q.size())
	assertHelper(t, &rb.q.root, rb.q.root.next, "when empty, expected queue.root.next equal to &queue.root")
	assertHelper(t, &rb.q.root, rb.q.root.prev, "when empty, expected queue.root.prev equal to &queue.root")
}

func TestPushAndPop(t *testing.T) {
	count := 5
	rb := New(count)

	for _, v := range []int{1, 3, 5} {
		rb.Push(v)
	}

	{
		// get 1 without error
		val, err := rb.Pop()
		assertHelper(t, val, 1, "expected to get 1, but got %v", val)
		assertHelper(t, nil, err, "unexpected error during pop: %v", err)
	}

	// push 4, [3, 5, 4]
	rb.Push(4)

	{
		// get 3 without error
		val, err := rb.Pop()
		assertHelper(t, val, 3, "expected to get 3, but got %v", val)
		assertHelper(t, nil, err, "unexpected error during pop: %v", err)
	}

	// push 2, [5, 4, 2]
	rb.Push(2)

	{
		// get 5 without error
		val, err := rb.Pop()
		assertHelper(t, val, 5, "expected to get 5, but got %v", val)
		assertHelper(t, nil, err, "unexpected error during pop: %v", err)
	}

	rb.Close()

	{
		// get error if push data into closed buffer
		_, err := rb.Push(0)
		assertHelper(t, ErrClosed, err,
			"expected to get error(%v) when push data into closed buffer, but got error(%v)", ErrClosed, err)
	}

	// check the buffer data
	expectedDump, got := []interface{}{4, 2}, rb.Drain()
	assertHelper(t, expectedDump, got, "expected return %v, but got %v", expectedDump, got)
}

func TestPopWaitWhenNotData(t *testing.T) {
	count := 5
	rb := New(count)

	var (
		wg     sync.WaitGroup
		waitCh = make(chan struct{}, 1)
	)

	wg.Add(1)
	go func() {
		waitCh <- struct{}{}

		defer wg.Done()
		_, rbErr := rb.Pop()
		close(waitCh)
		assertHelper(t, ErrClosed, rbErr,
			"expected to get error(%v) when push data into closed buffer, but got error(%v)", ErrClosed, rbErr)
	}()

	// make sure the goroutine has been scheduled
	<-waitCh
	select {
	case <-time.After(1 * time.Second):
		rb.Close()
		wg.Wait()
	case <-waitCh:
		t.Errorf("expect to block if there is no data in buffer")
		t.FailNow()
	}
}

func assertHelper(t *testing.T, expected, got interface{}, format string, args ...interface{}) {
	if !reflect.DeepEqual(expected, got) {
		t.Errorf(format, args...)
		t.FailNow()
	}
}
