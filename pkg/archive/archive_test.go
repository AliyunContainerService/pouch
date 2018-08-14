package archive

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alibaba/pouch/pkg/utils"
)

func TestCopyWithTar(t *testing.T) {
	source, err := ioutil.TempDir("", "source")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(source)

	destination, err := ioutil.TempDir("", "destination")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destination)

	files := []string{"file1", "file2", "dir1/file3", "dir2/file4"}

	err = makeFiles(source, files)
	if err != nil {
		t.Fatal(err)
	}

	err = CopyWithTar(source, destination)
	if err != nil {
		t.Fatal(err)
	}
	actualFiles := targetFiles(destination)

	if !utils.StringSliceEqual(files, actualFiles) {
		t.Fatalf(" TestCopyWithTar expected get %v, but got %v", files, actualFiles)
	}
}

func makeFiles(baseDir string, files []string) error {
	for _, file := range files {
		fullPath := path.Join(baseDir, file)

		dir := path.Dir(fullPath)

		// create dir.
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		// create file.
		fi, err := os.Create(fullPath)
		if err != nil {
			return err
		}
		fi.Close()
	}

	return nil
}

func targetFiles(baseDir string) []string {
	files := []string{}

	filepath.Walk(baseDir, func(name string, fi os.FileInfo, err error) error {
		if fi.Mode().IsRegular() {
			files = append(files, strings.TrimPrefix(name, baseDir+"/"))
		}
		return nil
	})

	return files
}
