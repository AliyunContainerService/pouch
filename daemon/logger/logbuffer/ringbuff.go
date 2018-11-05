package logbuffer

import (
	"fmt"
	"sync"

	"github.com/alibaba/pouch/daemon/logger"
)

// ErrClosed is used to indicate the ringbuffer has been closed.
var ErrClosed = fmt.Errorf("closed")

const (
	defaultMaxBytes = 1e6 //1MB
)

// RingBuffer implements a fixed-size buffer which will drop oldest data if full.
type RingBuffer struct {
	mu   sync.Mutex
	wait *sync.Cond

	closed bool
	q      *queue

	maxBytes     int64
	currentBytes int64
}

// NewRingBuffer creates new RingBuffer.
func NewRingBuffer(maxBytes int64) *RingBuffer {
	if maxBytes < 0 {
		maxBytes = defaultMaxBytes
	}

	rb := &RingBuffer{
		closed:   false,
		q:        newQueue(),
		maxBytes: maxBytes,
	}
	rb.wait = sync.NewCond(&rb.mu)
	return rb
}

// Push pushes value into buffer and return whether it covers the oldest data
// or not.
func (rb *RingBuffer) Push(val *logger.LogMessage) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.closed {
		return ErrClosed
	}

	if val == nil {
		return nil
	}

	msgLength := int64(len(val.Line))
	if (rb.currentBytes + msgLength) > rb.maxBytes {
		rb.wait.Broadcast()
		return nil
	}

	rb.q.enqueue(val)
	rb.wait.Broadcast()
	return nil
}

// Pop pops the value in the buffer.
//
// NOTE: it returns ErrClosed if the buffer has been closed.
func (rb *RingBuffer) Pop() (*logger.LogMessage, error) {
	rb.mu.Lock()
	for rb.q.size() == 0 && !rb.closed {
		rb.wait.Wait()
	}

	if rb.closed {
		rb.mu.Unlock()
		return nil, ErrClosed
	}

	val := rb.q.dequeue()
	rb.currentBytes -= int64(len(val.Line))
	rb.mu.Unlock()
	return val, nil
}

// Drain returns all the data in the buffer.
//
// NOTE: it can be used after closed to make sure the data have been consumed.
func (rb *RingBuffer) Drain() []*logger.LogMessage {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	size := rb.q.size()
	vals := make([]*logger.LogMessage, 0, size)

	for i := 0; i < size; i++ {
		vals = append(vals, rb.q.dequeue())
	}
	rb.currentBytes = 0
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
