package jsonfile

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/alibaba/pouch/daemon/logger"
)

func TestReadLogMessagesWithRemoveFileInFollowMode(t *testing.T) {
	f, err := ioutil.TempFile("", "tail-file")
	if err != nil {
		t.Fatalf("unexpected error during create tempfile: %v", err)
	}
	defer f.Close()
	defer os.RemoveAll(f.Name())

	jf, err := NewJSONLogFile(f.Name(), 0640, nil)
	if err != nil {
		t.Fatalf("unexpected error during create JSONLogFile: %v", err)
	}

	watcher := jf.ReadLogMessages(&logger.ReadConfig{Follow: true})
	defer watcher.Close()

	// NOTE: make the goroutine for read has started.
	<-time.After(100 * time.Millisecond)

	os.RemoveAll(f.Name())

	select {
	case _, _ = <-watcher.Msgs:
	case <-time.After(250 * time.Millisecond):
		// NOTE: watchTimeout is 200ms
		t.Fatal("expected watcher.Msgs has been closed after removed file, but it's still alive")
	}

	select {
	case err, ok := <-watcher.Err:
		t.Fatalf("unexpected error from watcher, but got %v %v", err, ok)
	default:
	}
}

func TestReadLogMessagesForEmptyFileWithoutFollow(t *testing.T) {
	f, err := ioutil.TempFile("", "tail-file")
	if err != nil {
		t.Fatalf("unexpected error during create tempfile: %v", err)
	}
	defer f.Close()
	defer os.RemoveAll(f.Name())

	jf, err := NewJSONLogFile(f.Name(), 0644, nil)
	if err != nil {
		t.Fatalf("unexpected error during create JSONLogFile: %v", err)
	}

	watcher := jf.ReadLogMessages(&logger.ReadConfig{})
	defer watcher.Close()

	// NOTE: make the goroutine for read has started.
	<-time.After(100 * time.Millisecond)

	select {
	case _, _ = <-watcher.Msgs:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("expected watcher.Msgs has been closed after removed file, but it's still alive")
	}

	select {
	case err, ok := <-watcher.Err:
		t.Fatalf("unexpected error from watcher, but got %v %v", err, ok)
	default:
	}
}
