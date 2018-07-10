package ctrd

import "testing"

func TestHasPort(t *testing.T) {
	// TODO
}

func TestCanonicalAddr(t *testing.T) {
	// TODO
}

func TestUseProxy(t *testing.T) {
	tests := []struct {
		name     string
		hostPort string
		wantErr  bool
	}{
		{name: "a", hostPort: "", wantErr: true},
		{name: "b", hostPort: "cannotBeSplit", wantErr: false},
		{name: "c", hostPort: "foo.com:80", wantErr: true},
		{name: "d", hostPort: "localhost:80", wantErr: false},
		{name: "e", hostPort: "192.168.0.1:80", wantErr: true},
		{name: "f", hostPort: ".foo.com:80", wantErr: true},
		{name: "f", hostPort: ".foo.com", wantErr: false},
	}
	//run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := useProxy(tt.hostPort)
			if result != tt.wantErr {
				t.Errorf("useProxy() = %v, want %v", result, tt.wantErr)
			}
		})
	}
}
