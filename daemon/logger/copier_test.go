package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"
)

type fakeJSONFileLogDriver struct {
	sync.Mutex
	*json.Encoder
}

func (ld *fakeJSONFileLogDriver) Name() string {
	return "fake-jsonfile"
}

func (ld *fakeJSONFileLogDriver) WriteLogMessage(msg *LogMessage) error {
	ld.Lock()
	defer ld.Unlock()
	return ld.Encode(msg)
}

func (ld *fakeJSONFileLogDriver) Close() error {
	return nil
}

func TestLogCopierBasic(t *testing.T) {
	stdoutContent := "data from testing process stdout\n"
	stderrContent := "data from testing process stderr\n"

	procStdout, procStderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	for i := 0; i < 100; i++ {
		if _, err := procStdout.Write([]byte(stdoutContent)); err != nil {
			t.Fatalf("failed to write data: %v", err)
		}
		if _, err := procStderr.Write([]byte(stderrContent)); err != nil {
			t.Fatalf("failed to write data: %v", err)
		}
	}

	jsonMsgBuf := bytes.NewBuffer(nil)
	lcopier := NewLogCopier(&fakeJSONFileLogDriver{Encoder: json.NewEncoder(jsonMsgBuf)},
		map[string]io.Reader{
			"stdout": procStdout,
			"stderr": procStderr,
		},
	)

	// NOTE: the bytes.Buffer.Read will return io.EOF when there is no data.
	lcopier.StartCopy()

	waitCh := make(chan struct{})
	go func() {
		lcopier.Wait()
		close(waitCh)
	}()
	select {
	case <-time.After(3 * time.Second):
		t.Fatal("take long time to finish copy")
	case <-waitCh:
	}

	// check the data
	dec := json.NewDecoder(jsonMsgBuf)
	for {
		var m LogMessage
		err := dec.Decode(&m)
		if err == io.EOF {
			return
		}

		if err != nil {
			t.Fatalf("failed to decode the json: %v", err)
		}

		switch m.Source {
		case "stdout":
			if got := string(m.Line); got != stdoutContent {
				t.Fatalf("[stdout] expected (%s), but got (%s)", stdoutContent, got)
			}
		case "stderr":
			if got := string(m.Line); got != stderrContent {
				t.Fatalf("[stderr] expected (%s), but got (%s)", stderrContent, got)
			}
		default:
			t.Fatalf("invalid the source type: %v", m.Source)
		}
	}
}
