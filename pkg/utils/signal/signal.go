package signal

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
)

const (
	sigrtmin = 34
	sigrtmax = 64
)

// SignalMap is a map of Linux signals.
var SignalMap = map[string]syscall.Signal{
	"ABRT":     syscall.SIGABRT,
	"ALRM":     syscall.SIGALRM,
	"BUS":      syscall.SIGBUS,
	"CHLD":     syscall.SIGCHLD,
	"CLD":      syscall.SIGCLD,
	"CONT":     syscall.SIGCONT,
	"FPE":      syscall.SIGFPE,
	"HUP":      syscall.SIGHUP,
	"ILL":      syscall.SIGILL,
	"INT":      syscall.SIGINT,
	"IO":       syscall.SIGIO,
	"IOT":      syscall.SIGIOT,
	"KILL":     syscall.SIGKILL,
	"PIPE":     syscall.SIGPIPE,
	"POLL":     syscall.SIGPOLL,
	"PROF":     syscall.SIGPROF,
	"PWR":      syscall.SIGPWR,
	"QUIT":     syscall.SIGQUIT,
	"SEGV":     syscall.SIGSEGV,
	"STKFLT":   syscall.SIGSTKFLT,
	"STOP":     syscall.SIGSTOP,
	"SYS":      syscall.SIGSYS,
	"TERM":     syscall.SIGTERM,
	"TRAP":     syscall.SIGTRAP,
	"TSTP":     syscall.SIGTSTP,
	"TTIN":     syscall.SIGTTIN,
	"TTOU":     syscall.SIGTTOU,
	"UNUSED":   syscall.SIGUNUSED,
	"URG":      syscall.SIGURG,
	"USR1":     syscall.SIGUSR1,
	"USR2":     syscall.SIGUSR2,
	"VTALRM":   syscall.SIGVTALRM,
	"WINCH":    syscall.SIGWINCH,
	"XCPU":     syscall.SIGXCPU,
	"XFSZ":     syscall.SIGXFSZ,
	"RTMIN":    sigrtmin,
	"RTMIN+1":  sigrtmin + 1,
	"RTMIN+2":  sigrtmin + 2,
	"RTMIN+3":  sigrtmin + 3,
	"RTMIN+4":  sigrtmin + 4,
	"RTMIN+5":  sigrtmin + 5,
	"RTMIN+6":  sigrtmin + 6,
	"RTMIN+7":  sigrtmin + 7,
	"RTMIN+8":  sigrtmin + 8,
	"RTMIN+9":  sigrtmin + 9,
	"RTMIN+10": sigrtmin + 10,
	"RTMIN+11": sigrtmin + 11,
	"RTMIN+12": sigrtmin + 12,
	"RTMIN+13": sigrtmin + 13,
	"RTMIN+14": sigrtmin + 14,
	"RTMIN+15": sigrtmin + 15,
	"RTMAX-14": sigrtmax - 14,
	"RTMAX-13": sigrtmax - 13,
	"RTMAX-12": sigrtmax - 12,
	"RTMAX-11": sigrtmax - 11,
	"RTMAX-10": sigrtmax - 10,
	"RTMAX-9":  sigrtmax - 9,
	"RTMAX-8":  sigrtmax - 8,
	"RTMAX-7":  sigrtmax - 7,
	"RTMAX-6":  sigrtmax - 6,
	"RTMAX-5":  sigrtmax - 5,
	"RTMAX-4":  sigrtmax - 4,
	"RTMAX-3":  sigrtmax - 3,
	"RTMAX-2":  sigrtmax - 2,
	"RTMAX-1":  sigrtmax - 1,
	"RTMAX":    sigrtmax,
}

// ParseSignal translates a string to a valid syscall signal.
// It returns an error if the signal map doesn't include the given signal.
func ParseSignal(rawSignal string) (syscall.Signal, error) {
	s, err := strconv.Atoi(rawSignal)
	if err == nil {
		// number is illegal
		if s <= 0 || s > sigrtmax {
			return -1, fmt.Errorf("Invalid signal: %s", rawSignal)
		}
		return syscall.Signal(s), nil
	}
	signal, ok := SignalMap[strings.TrimPrefix(strings.ToUpper(rawSignal), "SIG")]
	if !ok {
		return -1, fmt.Errorf("Invalid signal: %s", rawSignal)
	}
	return signal, nil
}
