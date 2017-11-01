package client

import "testing"

func TestNewClient(t *testing.T) {
	kvs := map[string]bool{
		"":                      false,
		"foobar":                true,
		"tcp://localhost:2476":  false,
		"http://localhost:2476": false,
	}

	for host, expectError := range kvs {
		cli, err := New(host)
		if err != nil {
			if !expectError {
				t.Fatalf("new client with host %v should not fail here %s", host, err)
			} else {
				t.Logf("new client with host %v should fail here: %s", host, err)
			}
		} else if err == nil && expectError {
			t.Fatalf("new client with host %v should fail here, but pass", host)
		}

		t.Logf("client info %+v", cli)
	}
}
