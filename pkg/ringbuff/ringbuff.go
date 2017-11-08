package ringbuff

import (
	"sync"
)

type ringNode struct {
	next  *ringNode
	value interface{}
}

// RingBuff implements a circular list.
type RingBuff struct {
	sync.Mutex
	cond    *sync.Cond
	pushPtr *ringNode
	popPtr  *ringNode
	closed  bool
}

// New creates a RingBuff.
func New(n int) *RingBuff {
	var (
		first *ringNode
		last  *ringNode
	)

	for i := 0; i < n; i++ {
		p := &ringNode{}

		if last != nil {
			last.next = p
		} else {
			first = p
		}

		last = p
	}

	last.next = first

	return &RingBuff{
		pushPtr: first,
		popPtr:  first,
		cond:    sync.NewCond(&sync.Mutex{}),
	}
}

// Push puts a elemnet into RingBuff and returns the status of covering or not.
func (r *RingBuff) Push(value interface{}) bool {
	r.Lock()
	defer r.Unlock()

	cover := false

	if r.closed {
		return cover
	}

	if r.pushPtr.value != nil {
		cover = true
	}

	// store value
	r.pushPtr.value = value

	// wakes the "Pop" goroutine waiting on "cond".
	if r.pushPtr == r.popPtr {
		r.cond.Broadcast()
	}

	// move pointer to next node.
	r.pushPtr = r.pushPtr.next

	return cover
}

// Pop returns a element, if RingBuff is empty, Pop() will block.
func (r *RingBuff) Pop() (interface{}, bool) {
	r.Lock()

	if v := r.popPtr.value; v != nil {
		isClosed := r.closed

		r.popPtr.value = nil // if we readed the node, must set nil to it.
		// move to next node.
		r.popPtr = r.popPtr.next

		// NOTICE: unlock
		r.Unlock()

		return v, isClosed
	}

	if r.closed {
		isClosed := r.closed

		// NOTICE: unlock
		r.Unlock()

		return nil, isClosed
	}

	// block util there is one element at least.
	r.cond.L.Lock()
	for r.popPtr.value == nil && !r.closed {
		// NOTICE: unlock, then to wait. if not call "Unlock", will block other's operation, eg: Push().
		r.Unlock()

		r.cond.Wait()

		// NOTICE: Wait() return, need to hold lock again.
		r.Lock()
	}
	r.cond.L.Unlock()

	v := r.popPtr.value
	isClosed := r.closed
	r.popPtr.value = nil // if we readed the node,  must set nil to it.
	// move to next node.
	r.popPtr = r.popPtr.next

	r.Unlock()

	return v, isClosed
}

// Close closes the RingBuff.
func (r *RingBuff) Close() error {
	r.Lock()
	r.closed = true
	r.cond.Broadcast()
	r.Unlock()
	return nil
}
