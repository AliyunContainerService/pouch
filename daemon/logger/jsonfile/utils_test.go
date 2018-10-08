package jsonfile

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/pkg/utils"
)

func generateFileBytes(lines int) []byte {
	buf := bytes.NewBuffer(nil)
	defer buf.Reset()

	for n := 1; n <= lines; n++ {
		buf.WriteString(fmt.Sprintf("#%d line\n", n))
	}
	return buf.Bytes()
}

func TestSeekOffsetByTailLines(t *testing.T) {
	f, err := ioutil.TempFile("", "tail-file")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.RemoveAll(f.Name())

	testContent := generateFileBytes(100)
	testContentInSlice := strings.Split(string(testContent), "\n")
	if _, err := f.Write(testContent); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		tail     int
		expected []string
	}{
		{
			tail:     100,
			expected: testContentInSlice[:len(testContentInSlice)-1],
		}, {
			tail:     2,
			expected: []string{"#99 line", "#100 line"},
		}, {
			tail:     1,
			expected: []string{"#100 line"},
		}, {
			tail:     1000,
			expected: testContentInSlice[:len(testContentInSlice)-1],
		}, {
			tail:     0,
			expected: testContentInSlice[:len(testContentInSlice)-1],
		},
	} {
		offset, err := seekOffsetByTailLines(f, tc.tail)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Seek(offset, os.SEEK_SET); err != nil {
			t.Fatal(err)
		}

		br := bufio.NewReader(f)
		for _, el := range tc.expected {
			l, _, err := br.ReadLine()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if el != string(l) {
				t.Fatalf("expected line %s, got %s", el, l)
			}
		}
	}
}

const (
	logContentPart1 = `{"stream":"stdout","log":"#1","time":"2018-05-09T10:00:01Z"}
{"stream":"stdout","log":"#2","time":"2018-05-09T10:00:02Z"}
{"stream":"stdout","log":"#3","time":"2018-05-09T10:00:03Z"}
`

	logContentPart2 = `{"stream":"stdout","log":"#4","time":"2018-05-09T10:00:04Z"}
`

	logContentPart3 = `{"stream":"stderr","log":"#5","time":"2018-05-09T10:00:05Z"}
{"stream":"stderr","log":"#6","time":"2018-05-09T10:00:06Z"}
`
)

func TestFollowFile(t *testing.T) {
	expectedMsgs := []*logger.LogMessage{
		{
			Source:    "stdout",
			Line:      []byte("#2"),
			Timestamp: generateTime(t, "2018-05-09T10:00:02Z"),
		},
		{
			Source:    "stdout",
			Line:      []byte("#3"),
			Timestamp: generateTime(t, "2018-05-09T10:00:03Z"),
		},
		{
			Source:    "stdout",
			Line:      []byte("#4"),
			Timestamp: generateTime(t, "2018-05-09T10:00:04Z"),
		},
	}

	// prepare the temp file
	f, err := ioutil.TempFile("", "tail-file")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.RemoveAll(f.Name())

	if _, err := f.Write([]byte(logContentPart1)); err != nil {
		t.Fatal(err)
	}

	// create goroutine to read the file
	watcher := logger.NewLogWatcher()
	waitCh := make(chan struct{})
	defer func() {
		watcher.Close()
		<-waitCh
	}()

	go func() {
		// NOTE: make sure all the goutine exits
		defer func() {
			waitCh <- struct{}{}
		}()

		newF, err := os.Open(f.Name())
		if err != nil {
			t.Fatalf("unexpected error during open file for followFile: %v", err)
		}

		cfg := &logger.ReadConfig{
			Since: generateTime(t, "2018-05-09T10:00:01.1Z"),
			Until: generateTime(t, "2018-05-09T10:00:04Z"),
		}
		followFile(newF, cfg, newUnmarshal, watcher)
	}()

	// should read three log message
	for _, el := range expectedMsgs[:2] {
		checkExpectedLogMessage(t, watcher, el)
	}

	// should be no more the log message
	expectedNoLogMessage(t, watcher)

	// write one more line
	if _, err := f.Write([]byte(logContentPart2)); err != nil {
		t.Fatal(err)
	}

	// should got the expected message
	checkExpectedLogMessage(t, watcher, expectedMsgs[2])

	// write two more lines
	if _, err := f.Write([]byte(logContentPart3)); err != nil {
		t.Fatal(err)
	}

	// should got the expected message because we set the until time
	expectedNoLogMessage(t, watcher)
}

func TestFailFile(t *testing.T) {
	logContent := []byte(logContentPart1)

	expectedMsgs := []*logger.LogMessage{
		{
			Source:    "stdout",
			Line:      []byte("#1"),
			Timestamp: generateTime(t, "2018-05-09T10:00:01Z"),
		},
		{
			Source:    "stdout",
			Line:      []byte("#2"),
			Timestamp: generateTime(t, "2018-05-09T10:00:02Z"),
		},
		{
			Source:    "stdout",
			Line:      []byte("#3"),
			Timestamp: generateTime(t, "2018-05-09T10:00:03Z"),
		},
	}

	watcher := logger.NewLogWatcher()

	for _, tc := range []struct {
		name     string
		cfg      *logger.ReadConfig
		expected []*logger.LogMessage
	}{
		{
			name:     "empty config",
			cfg:      &logger.ReadConfig{},
			expected: expectedMsgs,
		}, {
			name: "until 2018-05-09T10:00:06Z",
			cfg: &logger.ReadConfig{
				Until: generateTime(t, "2018-05-09T10:00:06Z"),
			},
			expected: expectedMsgs,
		}, {
			name: "until 2018-05-09T10:00:01Z",
			cfg: &logger.ReadConfig{
				Until: generateTime(t, "2018-05-09T10:00:01Z"),
			},
			expected: expectedMsgs[:1],
		}, {
			name: "since 2018-05-09T10:00:01.1Z",
			cfg: &logger.ReadConfig{
				Since: generateTime(t, "2018-05-09T10:00:01.1Z"),
			},
			expected: expectedMsgs[1:],
		}, {
			name: "since 2018-05-09T10:00:01.1Z, until 2018-05-09T10:00:02Z",
			cfg: &logger.ReadConfig{
				Since: generateTime(t, "2018-05-09T10:00:01.1Z"),
				Until: generateTime(t, "2018-05-09T10:00:02Z"),
			},
			expected: expectedMsgs[1:2],
		},
	} {
		{
			t.Run(tc.name, func(t *testing.T) {
				// NOTE: by default, the watcher has buffered chan for log message
				tailFile(bytes.NewBuffer(logContent), tc.cfg, newUnmarshal, watcher)

				for _, el := range tc.expected {
					checkExpectedLogMessage(t, watcher, el)
				}
				expectedNoLogMessage(t, watcher)
			})
		}
	}

}

func expectedNoLogMessage(t *testing.T, watcher *logger.LogWatcher) {
	select {
	case got, ok := <-watcher.Msgs:
		t.Fatalf("unexpected log message: %v, %v", got, ok)
	default:
	}
}

func checkExpectedLogMessage(t *testing.T, watcher *logger.LogWatcher, el *logger.LogMessage) {
	select {
	case got, ok := <-watcher.Msgs:
		if !ok {
			t.Fatal("unexpected close watcher.Msgs channel")
		}

		if !reflect.DeepEqual(el, got) {
			t.Fatalf("expected %v, but got %v", el, got)
		}
	case <-time.After(100 * time.Millisecond):
		// NOTE: make sure that the goroutine can load log in time.
		t.Fatal("expected log message here, but got nothing")
	}
}

func generateTime(t *testing.T, str string) time.Time {
	t1, err := time.Parse(utils.TimeLayout, str)
	if err != nil {
		t.Fatalf("unexpected error during parse time: %v", err)
	}
	return t1
}
