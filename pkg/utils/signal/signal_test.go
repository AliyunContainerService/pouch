package signal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSignal(t *testing.T) {
	type tCase struct {
		signal string
		value  int
		ok     bool
	}

	for _, tc := range []tCase{
		{
			signal: "SIGKILL",
			value:  9,
			ok:     true,
		},
		{
			signal: "SIGTERM",
			value:  15,
			ok:     true,
		},
		{
			signal: "SIGINT",
			value:  2,
			ok:     true,
		},
		{
			signal: "SIGQUIT",
			value:  3,
			ok:     true,
		},
		{
			signal: "SIGSTOP",
			value:  19,
			ok:     true,
		},
		{
			signal: "SIGCONT",
			value:  18,
			ok:     true,
		},
		{
			signal: "SIGUSR1",
			value:  10,
			ok:     true,
		},
		{
			signal: "SIGUSR2",
			value:  12,
			ok:     true,
		},
		{
			signal: "SIGCHLD",
			value:  17,
			ok:     true,
		},
		{
			signal: "SIGCONT",
			value:  18,
			ok:     true,
		},
		{
			signal: "STOP",
			value:  19,
			ok:     true,
		},
		{
			signal: "KILL",
			value:  9,
			ok:     true,
		},
		{
			signal: "123SIGTERM",
			ok:     false,
		},
		{
			signal: "23 INT",
			ok:     false,
		},
		{
			signal: "QUIT213",
			ok:     false,
		},
		{
			signal: "65",
			ok:     false,
		},
		{
			signal: "10",
			value:  10,
			ok:     true,
		},
	} {
		sig, _ := ParseSignal(tc.signal)
		if tc.ok {
			assert.Equal(t, tc.value, int(sig))
		} else {
			assert.Equal(t, -1, int(sig))
		}
	}
}
