package v1alpha2

import (
	"testing"
)

func Test_extractIPAndPortFromAddresses(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantIP   string
		wantPort string
	}{
		{
			name:     "listening addresses are nil",
			args:     nil,
			wantIP:   "",
			wantPort: "",
		},
		{
			name:     "listening addresses have no tcp address",
			args:     []string{"unix:///var/run/pouchd.sock"},
			wantIP:   "",
			wantPort: "",
		},
		{
			name:     "listening addresses have valid address",
			args:     []string{"unix:///var/run/pouchd.sock", "tcp://0.0.0.0:4345"},
			wantIP:   "0.0.0.0",
			wantPort: "4345",
		},
		{
			name:     "listening addresses have two tcp addresses",
			args:     []string{"tcp://10.10.10.10:1234", "tcp://0.0.0.0:4345"},
			wantIP:   "10.10.10.10",
			wantPort: "1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIP, gotPort := extractIPAndPortFromAddresses(tt.args)
			if gotIP != tt.wantIP {
				t.Errorf("extractIPAndPortFromAddresses() IP = %v, want IP %v", gotIP, tt.wantIP)
			}
			if gotPort != tt.wantPort {
				t.Errorf("extractIPAndPortFromAddresses() Port = %v, want Port %v", gotPort, tt.wantPort)
			}
		})
	}
}
