package stream

import (
	"testing"
)

// TODO: use fake clock to test gc of request cache.

func TestRequestCacheBasic(t *testing.T) {
	var tokens []string
	r := NewRequestCache()

	n := 10
	for i := 0; i < n; i++ {
		token, err := r.Insert(i)
		if err != nil {
			t.Fatalf("unexpected error when inserting the request: %v", err)
		}
		tokens = append(tokens, token)
	}

	for i := 0; i < n; i++ {
		req, found := r.Consume(tokens[i])
		if !found {
			t.Fatalf("unexpected error when comsuming the cached request")
		}
		r, ok := req.(int)
		if !ok {
			t.Fatalf("the type of cached request has been changed")
		}
		if r != i {
			t.Fatalf("the value of cached request has been changed")
		}
	}
}

func TestRequestCacheNonExist(t *testing.T) {
	r := NewRequestCache()
	token := "non-exist"
	_, found := r.Consume(token)
	if found {
		t.Fatalf("should not find the request that not exist")
	}
}

func TestRequestCacheTokenUnique(t *testing.T) {
	r := NewRequestCache()
	tokens := make(map[string]bool)
	for i := 0; i < MaxInFlight; i++ {
		token, err := r.Insert(i)
		if err != nil {
			t.Fatalf("unexpected error when inserting the request: %v", err)
		}
		if tokens[token] {
			t.Fatalf("the returned token should be unique")
		}
		tokens[token] = true
	}
}

func TestRequestCacheMaxInFlight(t *testing.T) {
	r := NewRequestCache()

	var i int
	for i = 0; i < MaxInFlight; i++ {
		_, err := r.Insert(i)
		if err != nil {
			t.Fatalf("unexpected error when inserting the request: %v", err)
		}
	}
	_, err := r.Insert(i)
	if err == nil {
		t.Fatalf("should report error when there are too many cached request")
	}
}
