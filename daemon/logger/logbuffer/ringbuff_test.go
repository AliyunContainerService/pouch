package logbuffer

import (
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/alibaba/pouch/daemon/logger"
)

func TestPushNormal(t *testing.T) {
	rb := NewRingBuffer(1024)

	b := make([]byte, 1024)
	extraB := []byte{1}

	// push bytes of max size
	err := rb.Push(wrapLogWithByte(b))
	assertHelper(t, nil, err, "unexpected error during push non-closed queue: %v", err)

	// continue to push new data
	err = rb.Push(wrapLogWithByte(extraB))
	assertHelper(t, nil, err, "unexpected error during push non-closed queue: %v", err)

	// get data
	logMsg, err := rb.Pop()
	expectedDump := wrapLogWithByte(b)
	assertHelper(t, nil, err, "unexpected error during pop: %v", err)
	assertHelper(t, expectedDump, logMsg, "expected return %v, but got %v", expectedDump, logMsg)

	// get drain data
	got := rb.Drain()
	expectedLogs := []*logger.LogMessage{wrapLogWithByte(extraB)}
	assertHelper(t, expectedLogs, got, "expected return %v, but got %v", expectedLogs, got)

	assertHelper(t, 0, rb.q.size(), "expected to have empty queue, but got %d size of queue", rb.q.size())
	assertHelper(t, &rb.q.root, rb.q.root.next, "when empty, expected queue.root.next equal to &queue.root")
	assertHelper(t, &rb.q.root, rb.q.root.prev, "when empty, expected queue.root.prev equal to &queue.root")

}

func TestPushAndPop(t *testing.T) {
	rb := NewRingBuffer(defaultMaxBytes)

	for _, v := range []int{1, 3, 5} {
		rb.Push(wrapLogWithInt(v))
	}

	{
		// get 1 without error
		val, err := rb.Pop()
		assertHelper(t, val, wrapLogWithInt(1), "expected to get 1, but got %v", val)
		assertHelper(t, nil, err, "unexpected error during pop: %v", err)
	}

	// push 4, [3, 5, 4]
	rb.Push(wrapLogWithInt(4))

	{
		// get 3 without error
		val, err := rb.Pop()
		assertHelper(t, val, wrapLogWithInt(3), "expected to get 3, but got %v", val)
		assertHelper(t, nil, err, "unexpected error during pop: %v", err)
	}

	// push 2, [5, 4, 2]
	rb.Push(wrapLogWithInt(2))

	{
		// get 5 without error
		val, err := rb.Pop()
		assertHelper(t, val, wrapLogWithInt(5), "expected to get 5, but got %v", val)
		assertHelper(t, nil, err, "unexpected error during pop: %v", err)
	}

	rb.Close()

	{
		// get error if push data into closed buffer
		err := rb.Push(wrapLogWithInt(0))
		assertHelper(t, ErrClosed, err,
			"expected to get error(%v) when push data into closed buffer, but got error(%v)", ErrClosed, err)
	}

	// check the buffer data
	expectedDump, got := []*logger.LogMessage{wrapLogWithInt(4), wrapLogWithInt(2)}, rb.Drain()
	assertHelper(t, expectedDump, got, "expected return %v, but got %v", expectedDump, got)
}

func TestPopWaitWhenNotData(t *testing.T) {
	rb := NewRingBuffer(defaultMaxBytes)

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

func wrapLogWithInt(num int) *logger.LogMessage {
	return &logger.LogMessage{
		Line: []byte(strconv.Itoa(num)),
	}
}

func wrapLogWithByte(bytes []byte) *logger.LogMessage {
	return &logger.LogMessage{
		Line: bytes,
	}
}
