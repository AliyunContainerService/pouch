package streams

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

type badWriter struct{}

func (w *badWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("oops")
}

func (w *badWriter) Close() error {
	return nil
}

type bufferWrapper struct {
	*bytes.Buffer
}

func (w *bufferWrapper) Close() error {
	return nil
}

func TestMultiWriter(t *testing.T) {
	mw := new(multiWriter)

	var (
		w1 = &badWriter{}
		w2 = &bufferWrapper{bytes.NewBuffer(nil)}
		w3 = &badWriter{}
	)

	mw.Add(w1)
	mw.Add(w2)
	mw.Add(w3)

	// step1: write hello
	n, err := mw.Write([]byte("hello"))
	if n != 5 || err != nil {
		t.Fatalf("failed to write data: n=%v, err=%v", n, err)
	}

	if len(mw.writers) != 1 || !reflect.DeepEqual(mw.writers[0], w2) {
		t.Fatal("failed to evict the bad writer")
	}

	// step2: write pouch
	n, err = mw.Write([]byte("pouch"))
	if n != 5 || err != nil {
		t.Fatalf("failed to write data: n=%v, err=%v", n, err)
	}

	if w2.String() != "hellopouch" {
		t.Fatalf("failed to write data")
	}

	// step3: close
	mw.Close()
	if len(mw.writers) != 0 {
		t.Fatal("failed to remove all the writers after close")
	}
}
