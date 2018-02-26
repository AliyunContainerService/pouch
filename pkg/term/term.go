package term

import (
	"os"
	"syscall"
	"unsafe"
)

// StdinEcho enable or disable echoing standard terminal input.
func StdinEcho(echo bool) error {
	return TerminalEcho(os.Stdin.Fd(), echo)
}

// StdoutEcho enable or disable echoing standard terminal output.
func StdoutEcho(echo bool) error {
	return TerminalEcho(os.Stdout.Fd(), echo)
}

// TerminalRestore restores terminal state connected to the file descriptor with the specific termios.
func TerminalRestore(fd uintptr, termios *syscall.Termios) error {
	return tcset(fd, termios)
}

// TerminalEcho enable or disable echoing terminal put which connected to the given file descriptor.
func TerminalEcho(fd uintptr, echo bool) error {
	termios := &syscall.Termios{}
	if err := tcget(fd, termios); err != 0 {
		return err
	}

	if echo {
		termios.Lflag |= syscall.ECHO
	} else {
		termios.Lflag &^= syscall.ECHO
	}

	return tcset(fd, termios)
}

func tcget(fd uintptr, termios *syscall.Termios) syscall.Errno {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(termios)))
	return err
}

func tcset(fd uintptr, termios *syscall.Termios) syscall.Errno {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(termios)))
	return err
}
