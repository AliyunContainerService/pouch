package util

import (
	"strconv"
	"strings"
)

const goroutine = "goroutine"

func isdigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// Goroutine represents the goroutine from runtime.Stack.
type Goroutine struct {
	ID         int64
	Stacktrace string
}

// GoroutinesFromStack returns the Goroutine by the stack information.
func GoroutinesFromStack(stacks []byte) []*Goroutine {
	bufs := strings.Split(string(stacks), "\n\n")
	res := make([]*Goroutine, 0, len(bufs))

	for _, buf := range bufs {
		bs := strings.SplitN(buf, "\n", 2)
		res = append(res, &Goroutine{
			ID:         parseGoID(bs[0]),
			Stacktrace: bs[1],
		})
	}
	return res
}

func parseGoID(buf string) int64 {
	idx := 0

	buf = buf[len(goroutine)+1:]
	for idx < len(buf) && isdigit(buf[idx]) {
		idx++
	}

	id, _ := strconv.ParseInt(string(buf[:idx]), 10, 64)
	return id
}
