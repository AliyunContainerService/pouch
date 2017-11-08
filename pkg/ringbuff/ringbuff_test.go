package ringbuff

import (
	"testing"
	"time"
)

func TestPushRewrite(t *testing.T) {
	ring := New(10)

	for i := 0; i < 10; i++ {
		if rewrite := ring.Push(i); rewrite {
			t.Fatalf("the ring buffer's size is error")
		}
	}

	for i := 0; i < 10; i++ {
		if rewrite := ring.Push(i); !rewrite {
			t.Fatalf("don't rewrite the node")
		}
	}
}

func TestPopBlock(t *testing.T) {
	ring := New(10)

	wait := make(chan struct{})
	go func() {
		ring.Pop()
		close(wait)
	}()

	select {
	case <-time.After(time.Second * 5):
	case <-wait:
		t.Errorf("not to block")
	}
}

func TestPushPop(t *testing.T) {
	ring := New(10)

	for i := 0; i < 10; i++ {
		ring.Push(i)
	}

	for i := 0; i < 10; i++ {
		v, _ := ring.Pop()
		if v.(int) != i {
			t.Errorf("failed to pop, <%d, %d>", v.(int), i)
		}
	}

	go func() {
		time.Sleep(time.Second * 5)
		ring.Push(111)
	}()

	v, _ := ring.Pop()
	if v.(int) != 111 {
		t.Errorf("failed to pop, <%d, %d>", v.(int), 111)
	}
}
