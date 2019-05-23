package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func tmpFileWithContent(content string, t *testing.T) string {
	tmpFile, err := ioutil.TempFile("", "envfile-test")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	tmpFile.WriteString(content)
	return tmpFile.Name()
}

// Test ParseEnvFile for a file with a few well formatted lines
func TestParseEnvFileGoodFile(t *testing.T) {
	content := `foo=bar
    baz=quux
# comment

_foobar=foobaz
with.dots=working
and_underscore=working too
`
	// Adding a newline + a line with pure whitespace.
	// This is being done like this instead of the block above
	// because it's common for editors to trim trailing whitespace
	// from lines, which becomes annoying since that's the
	// exact thing we need to test.
	content += "\n    \t  "
	tmpFile := tmpFileWithContent(content, t)
	defer os.Remove(tmpFile)

	lines, err := ParseEnvFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	expectedLines := []string{
		"foo=bar",
		"baz=quux",
		"_foobar=foobaz",
		"with.dots=working",
		"and_underscore=working too",
	}

	if !reflect.DeepEqual(lines, expectedLines) {
		t.Fatal("lines not equal to expected_lines")
	}
}
