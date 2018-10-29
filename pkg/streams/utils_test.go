package streams

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func TestAttachWithCloseStdin(t *testing.T) {
	var (
		aStdin  = &bufferWrapper{bytes.NewBuffer(nil)}
		aStdout = bytes.NewBuffer(nil)
		aStderr = bytes.NewBuffer(nil)
	)

	attachCfg := &AttachConfig{
		UseStdin:   true,
		Stdin:      aStdin,
		UseStdout:  true,
		Stdout:     aStdout,
		UseStderr:  false,
		Stderr:     aStderr,
		CloseStdin: true, // must set it to true
	}

	stream := NewStream()
	stream.NewStdinInput()

	// write data into the aStdin
	content := ""
	for i := 0; i < 100; i++ {
		d := "hello"
		if _, err := aStdin.Write([]byte(d)); err != nil {
			t.Fatalf("failed to write data: %v", err)
		}
		content += d
	}

	// start attach stream
	attachErr := stream.Attach(context.Background(), attachCfg)

	// write data into stderr, but the data never goes into the aStderr
	stream.Stderr().Write([]byte("hello stderr"))

	// read from stdin and echo it into stdout
	echoRout, echoW := io.Pipe()
	go func() {
		io.Copy(echoW, stream.Stdin())
		echoW.Close()
	}()

	go func() {
		io.Copy(stream.Stdout(), echoRout)
		stream.Stdout().Close()
	}()

	aStdin.Close()
	if err := <-attachErr; err != nil {
		t.Fatalf("failed to attach: %v", err)
	}

	if got := aStdout.String(); got != content {
		t.Fatalf("expected to get (%s), but got (%s)", content, got)
	}

	// UseStderr is false
	if aStderr.String() != "" {
		t.Fatalf("should not get any data in stderr, but got %v", aStderr.String())
	}

	if err := stream.Close(); err != nil {
		t.Fatalf("failed to stop stream: %v", err)
	}
}
