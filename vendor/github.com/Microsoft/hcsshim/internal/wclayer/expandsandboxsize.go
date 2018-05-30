package wclayer

import (
	"github.com/Microsoft/hcsshim/internal/hcserror"
	"github.com/sirupsen/logrus"
)

// ExpandSandboxSize expands the size of a layer to at least size bytes.
func ExpandSandboxSize(path string, size uint64) error {
	title := "hcsshim::ExpandSandboxSize "
	logrus.Debugf(title+"path=%s size=%d", path, size)

	err := expandSandboxSize(&stdDriverInfo, path, size)
	if err != nil {
		err = hcserror.Errorf(err, title, "path=%s size=%d", path, size)
		logrus.Error(err)
		return err
	}

	logrus.Debugf(title+"- succeeded path=%s size=%d", path, size)
	return nil
}
