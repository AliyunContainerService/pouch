// +build linux

package system

import (
	"syscall"

	"github.com/pkg/errors"
)

// GetDevID returns device id via syscall according to the input directory.
func GetDevID(dir string) (uint64, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(dir, &st); err != nil {
		return 0, errors.Wrapf(err, "failed to get device id of directory: (%s)", dir)
	}
	return st.Dev, nil
}
