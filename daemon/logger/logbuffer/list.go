package logbuffer

import (
	"sync"

	"github.com/alibaba/pouch/daemon/logger"
)

var elemPool = &sync.Pool{New: func() interface{} { return new(element) }}

type element struct {
	next, prev *element
	val        *logger.LogMessage
}

func (e *element) reset() {
	e.next, e.prev = nil, nil
	e.val = nil
}

type queue struct {
	root  element
	count int
}

func newQueue() *queue {
	q := new(queue)

	q.root.next = &q.root
	q.root.prev = &q.root
	q.count = 0
	return q
}

func (q *queue) size() int {
	return q.count
}

func (q *queue) enqueue(val *logger.LogMessage) {
	elem := elemPool.Get().(*element)
	elem.val = val

	at := q.root.prev

	at.next = elem
	elem.prev = at
	elem.next = &q.root
	q.root.prev = elem
	q.count++
}

func (q *queue) dequeue() *logger.LogMessage {
	if q.size() == 0 {
		return nil
	}

	at := q.root.next
	at.prev.next = at.next
	at.next.prev = at.prev
	val := at.val

	at.reset()
	elemPool.Put(at)
	q.count--
	return val
}
