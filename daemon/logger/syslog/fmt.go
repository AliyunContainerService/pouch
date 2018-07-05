package syslog

import (
	"fmt"
	"os"
	"time"

	"github.com/RackSec/srslog"
)

func rfc5424FormatterWithTagAsAppName(p srslog.Priority, hostname, tag, content string) string {
	timestamp := time.Now().Format(time.RFC3339)
	pid := os.Getpid()
	msg := fmt.Sprintf("<%d>%d %s %s %s %d %s - %s",
		p, 1, timestamp, hostname, tag, pid, tag, content)
	return msg
}

func rfc5424MicroFormatterWithTagAsAppName(p srslog.Priority, hostname, tag, content string) string {
	timestamp := time.Now().Format(timeRfc5424fmt)
	pid := os.Getpid()
	msg := fmt.Sprintf("<%d>%d %s %s %s %d %s - %s",
		p, 1, timestamp, hostname, tag, pid, tag, content)
	return msg
}

func rfc5424MicroFormatterWithSequenceAndTagAsAppName(p srslog.Priority, hostname, tag, content string) string {
	now := time.Now()
	pid := os.Getpid()
	msg := fmt.Sprintf("<%d>%d %d %s %s %s %d %s - %s",
		p, 1, now.UnixNano(), now.Format(timeRfc5424fmt), hostname, tag, pid, tag, content)
	return msg
}
