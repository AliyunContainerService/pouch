package stream

import (
	"testing"
)

// TODO: use fake clock to test gc of request cache.

func TestRequestCacheBasic(t *testing.T) {
	var tokens []string
	r := NewRequestCache()
	
	for i := 0; i < 10; i++ {
		token, err := r.Insert(i)
		if err != nil {
			t.Fatalf("unexpected error when inserting the request: %v", err)
		}
		tokens = append(tokens, token)
	}

	for i := 0; i < 10; i++ {
		require, found := r.Consume(tokens[i])
		if !found {
			t.Fatalf("unexpected error when comsuming the cached request")
		}
		r, ok := require.(int)
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
