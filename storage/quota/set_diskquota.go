package quota

import (
	"fmt"
	"os"
	"strconv"

	"github.com/docker/docker/pkg/reexec"
	"github.com/sirupsen/logrus"
)

func init() {
	reexec.Register("set-diskquota", processSetQuotaReexec)
}

// OverlayMount represents the parameters of overlay mount.
type OverlayMount struct {
	Merged string
	Lower  string
	Upper  string
	Work   string
}

func processSetQuotaReexec() {
	var (
		err error
		qid uint32
	)

	// Return a failure to the calling process via ExitCode
	defer func() {
		if err != nil {
			logrus.Fatalf("%v", err)
		}
	}()

	if len(os.Args) != 4 {
		err = fmt.Errorf("invalid arguments: %v, it should be: %s: <path> <size> <quota id>", os.Args, os.Args[0])
		return
	}

	basefs := os.Args[1]
	size := os.Args[2]
	id, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return
	}
	qid = uint32(id)

	logrus.Infof("set diskquota: %v", os.Args)

	err = SetRootfsDiskQuota(basefs, size, qid)

	return
}
