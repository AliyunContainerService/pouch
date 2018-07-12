// +build linux

package system

import (
	"syscall"

	"github.com/sirupsen/logrus"
)

// GetDevID returns device id via syscall according to the input directory.
func GetDevID(dir string) (uint64, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(dir, &st); err != nil {
		logrus.Warnf("failed to get device id of dir %s: %v", dir, err)
		return 0, err
	}
	return st.Dev, nil
}
