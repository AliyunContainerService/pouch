package ringbuffer

import (
	"fmt"
	"sync"
)

// ErrClosed is used to indicate the ringbuffer has been closed.
var ErrClosed = fmt.Errorf("closed")

const defaultSize = 1024

// RingBuffer implements a fixed-size buffer which will drop oldest data if full.
type RingBuffer struct {
	mu   sync.Mutex
	wait *sync.Cond

	cap    int
	closed bool
	q      *queue
}

// New creates new RingBuffer.
func New(cap int) *RingBuffer {
	if cap <= 0 {
		cap = defaultSize
	}

	rb := &RingBuffer{
		cap:    cap,
		closed: false,
		q:      newQueue(),
	}
	rb.wait = sync.NewCond(&rb.mu)
	return rb
}

// Push pushes value into buffer and return whether it covers the oldest data
// or not.
func (rb *RingBuffer) Push(val interface{}) (bool, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.closed {
		return false, ErrClosed
	}

	if val == nil {
		return false, nil
	}

	// drop the oldest element if covered
	covered := (rb.q.size() == rb.cap)
	if covered {
		rb.q.dequeue()
	}

	rb.q.enqueue(val)
	rb.wait.Broadcast()
	return covered, nil
}

// Pop pops the value in the buffer.
//
// NOTE: it returns ErrClosed if the buffer has been closed.
func (rb *RingBuffer) Pop() (interface{}, error) {
	rb.mu.Lock()
	for rb.q.size() == 0 && !rb.closed {
		rb.wait.Wait()
	}

	if rb.closed {
		rb.mu.Unlock()
		return nil, ErrClosed
	}

	val := rb.q.dequeue()
	rb.mu.Unlock()
	return val, nil
}

// Drain returns all the data in the buffer.
//
// NOTE: it can be used after closed to make sure the data have been consumed.
func (rb *RingBuffer) Drain() []interface{} {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	size := rb.q.size()
	vals := make([]interface{}, 0, size)

	for i := 0; i < size; i++ {
		vals = append(vals, rb.q.dequeue())
	}
	return vals
}

// Close closes the ringbuffer.
func (rb *RingBuffer) Close() error {
	rb.mu.Lock()
	if rb.closed {
		rb.mu.Unlock()
		return nil
	}

	rb.closed = true
	rb.wait.Broadcast()
	rb.mu.Unlock()
	return nil
}
