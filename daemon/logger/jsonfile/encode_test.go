package jsonfile

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/alibaba/pouch/daemon/logger"
)

func TestMarshalAndUnmarshal(t *testing.T) {

	attrs := map[string]string{"env": "test"}
	extra, err := json.Marshal(attrs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedMsg := &logger.LogMessage{
		Source:    "stdout",
		Line:      []byte("hello pouch"),
		Timestamp: time.Now().UTC(),
		Attrs:     attrs,
	}

	bs, err := Marshal(expectedMsg, extra)
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
