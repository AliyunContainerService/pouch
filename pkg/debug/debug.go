package debug

import (
	"os"
	"os/signal"
	"runtime"
	rdebug "runtime/debug"
	"syscall"

	"github.com/alibaba/pouch/pkg/log"
)

func init() {
	rdebug.SetTraceback("all")
}

// SetupDumpStackTrap setups signal trap to dump stack.
func SetupDumpStackTrap() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	go func() {
		for range c {
			DumpStacks()
		}
	}()
}

// DumpStacks dumps the runtime stack.
func DumpStacks() {
	var (
		buf       []byte
		stackSize int
	)
	bufferLen := 16384
	for stackSize == len(buf) {
		buf = make([]byte, bufferLen)
		stackSize = runtime.Stack(buf, true)
		bufferLen *= 2
	}
	buf = buf[:stackSize]
	// Note that if the daemon is started with a less-verbose log-level than "info" (the default), the goroutine
	// traces won't show up in the log.
	log.With(nil).Infof("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
}
