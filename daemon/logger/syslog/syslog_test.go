package syslog

import (
	"testing"

	"github.com/alibaba/pouch/daemon/logger"

	"github.com/RackSec/srslog"
)

func TestParseOptions(t *testing.T) {
	info := logger.Info{
		LogConfig: map[string]string{
			"syslog-address":  "tcp://localhost:8080",
			"syslog-facility": "daemon",
			"syslog-format":   "rfc3164",
			"tag":             "{{.FullID}}",
		},
		ContainerID: "container-20180610",
	}

	opts, err := parseOptions(info)
	if err != nil {
		t.Fatalf("failed to parse options: %v", err)
	}

	// check formatter and framer
	if !isSameFunc(opts.formatter, srslog.RFC3164Formatter) || !isSameFunc(opts.framer, srslog.DefaultFramer) {
		t.Fatalf("expect formatter(%v) & framer(%v), but got formatter(%v) & framer(%v)",
			srslog.RFC3164Formatter, srslog.DefaultFramer, opts.formatter, opts.framer,
		)
	}

	// check tag
	if expected := info.ContainerID; opts.tag != expected {
		t.Fatalf("expect tag(%v), but got tag(%v)", expected, opts.tag)
	}

	// check priority
	if expected := srslog.LOG_DAEMON; opts.priority != expected {
		t.Fatalf("expect priority(%v), but got priority(%v)", expected, opts.priority)
	}

	// check proto and address
	if proto, addr := "tcp", "localhost:8080"; opts.proto != proto || opts.address != addr {
		t.Fatalf("expect proto(%v) & address(%v), but got proto(%v) & address(%v)",
			proto, addr, opts.proto, opts.address,
		)
	}
}
