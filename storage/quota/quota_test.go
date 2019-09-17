// +build linux

package quota

import (
	"os"
	"testing"

	"github.com/alibaba/pouch/pkg/system"
)

func Test_getDevID(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get work directory error %v", err)
	}
	expectID, err := system.GetDevID(wd)
	if err != nil {
		t.Fatalf("get dev id error of %s: %v", wd, err)
	}

	gotID, err := getDevID(wd)
	if err != nil {
		t.Fatalf("get dev id error of %s by getDevID: %v", wd, err)
	}

	if expectID != gotID {
		t.Fatalf("getDevID error expect %d got %d", expectID, gotID)
	}
}
