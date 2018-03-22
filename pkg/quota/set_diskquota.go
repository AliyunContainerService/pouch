package quota

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

	if len(os.Args) != 3 {
		err = fmt.Errorf("invalid arguments: %v, it should be: %s: <path> <size>", os.Args, os.Args[0])
		return
	}

	basefs := os.Args[1]
	size := os.Args[2]

	logrus.Infof("set diskquota: %v", os.Args)

	overlayfs, err := getOverlay(basefs)
	if err != nil || overlayfs == nil {
		logrus.Errorf("failed to get lowerdir: %v", err)
		return
	}

	for _, dir := range []string{overlayfs.Upper, overlayfs.Work} {
		_, err = StartQuotaDriver(dir)
		if err != nil {
			logrus.Errorf("failed to start quota driver: %v", err)
			return
		}

		qid, err = SetSubtree(dir, qid)
		if err != nil {
			logrus.Errorf("failed to set subtree: %v", err)
			return
		}

		err = SetDiskQuota(dir, size, int(qid))
		if err != nil {
			logrus.Errorf("failed to set disk quota: %v", err)
		}

		setQuotaForDir(dir, qid)
	}

	return
}

func setQuotaForDir(src string, qid uint32) {
	filepath.Walk(src, func(path string, fd os.FileInfo, err error) error {
		if err != nil {
			logrus.Warnf("setQuota walk dir %s get error %v", path, err)
			return nil
		}
		SetFileAttrNoOutput(path, qid)
		return nil
	})
}

func getOverlay(basefs string) (*OverlayMount, error) {
	overlayfs := &OverlayMount{}

	fd, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	br := bufio.NewReader(fd)
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}

		parts := strings.Split(string(line), " ")
		if len(parts) != 6 {
			continue
		}
		if parts[1] != basefs || parts[2] != "overlay" {
			continue
		}

		mountParams := strings.Split(parts[3], ",")
		for _, p := range mountParams {
			switch {
			case strings.Contains(p, "lowerdir"):
				if s := strings.Split(p, "="); len(s) == 2 {
					overlayfs.Lower = s[1]
				}

			case strings.Contains(p, "upperdir"):
				if s := strings.Split(p, "="); len(s) == 2 {
					overlayfs.Upper = s[1]
				}

			case strings.Contains(p, "workdir"):
				if s := strings.Split(p, "="); len(s) == 2 {
					overlayfs.Work = s[1]
					break
				}
			}
		}
	}

	return overlayfs, nil
}
