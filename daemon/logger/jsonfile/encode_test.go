package jsonfile

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/alibaba/pouch/daemon/logger"
)

func TestMarshalAndUnmarshal(t *testing.T) {
	expectedMsg := &logger.LogMessage{
		Source:    "stdout",
		Line:      []byte("hello pouch"),
		Timestamp: time.Now().UTC(),
	}

	bs, err := marshal(expectedMsg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	decodeOneLine := newUnmarshal(bytes.NewBuffer(bs))
	got, err := decodeOneLine()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expectedMsg, got) {
		t.Fatalf("expected %v, but got %v", expectedMsg, got)
	}
}
