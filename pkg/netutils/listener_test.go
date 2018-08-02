package netutils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetListenerBasic(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			"tcpAddressTest",
			args{"tcp://127.0.0.1:12345"},
			nil,
		},
		{
			"unixAddressTest",
			args{"unix:///tmp/pouchtest.sock"},
			nil,
		},
		{
			"otherProtocolTest",
			args{"udp://127.0.0.1:12345"},
			fmt.Errorf("only unix socket or tcp address is support"),
		},
		{
			"invalidAddressTest",
			args{"invalid address"},
			fmt.Errorf("invalid listening address invalid address: must be in format [protocol]://[address]"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := GetListener(tt.args.addr, nil)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetListener() return error %v, want %v", err, tt.wantErr)
			}
			if err == nil {
				l.Close()
			}
		})
	}
}
