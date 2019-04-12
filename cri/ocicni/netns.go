package ocicni

import (
	"crypto/rand"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

const nsRunDir = "/var/run/netns"

// NewNetNS creates a new persistent network namespace and returns the
// namespace path, without switching to it
func (c *CniManager) NewNetNS() (string, error) {
	return createNS("")
}

// RemoveNetNS unmounts the network namespace
func (c *CniManager) RemoveNetNS(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "failed to stat netns")
	}
	if strings.HasPrefix(path, nsRunDir) {
		if err := unix.Unmount(path, 0); err != nil {
			return errors.Wrapf(err, "failed to unmount NS: at %s", path)
		}

		if err := os.Remove(path); err != nil {
			return errors.Wrapf(err, "failed to remove ns path %s", path)
		}
	}

	return nil
}

// CloseNetNS cleans up this instance of the network namespace; if this instance
// is the last user the namespace will be destroyed
func (c *CniManager) CloseNetNS(path string) error {
	netns, err := ns.GetNS(path)

	if err != nil {
		if _, ok := err.(ns.NSPathNotExistErr); ok {
			return nil
		}
		if _, ok := err.(ns.NSPathNotNSErr); ok {
			if err := os.RemoveAll(path); err != nil {
				return errors.Wrapf(err, "failed to remove netns path %s", path)
			}
			return nil
		}
		return errors.Wrapf(err, "failed to get netns path %s", path)
	}
	if err := netns.Close(); err != nil {
		return errors.Wrapf(err, " failed to clean up netns path %s", path)
	}
	return nil
}

// RecoverNetNS recreate a persistent network namespace if the ns is not exists.
// Otherwise, do nothing.
func (c *CniManager) RecoverNetNS(path string) error {
	_, err := ns.GetNS(path)

	// net ns already exists
	if err == nil {
		return nil
	}

	_, err = createNS(path)
	return err
}

// getCurrentThreadNetNSPath copied from pkg/ns
func getCurrentThreadNetNSPath() string {
	// /proc/self/ns/net returns the namespace of the main thread, not
	// of whatever thread this goroutine is running on.  Make sure we
	// use the thread's net namespace since the thread is switching around
	return fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), unix.Gettid())
}

// createNS create and mount the network namespace of the given path, or create a brand new one.
// partially copy some code from https://github.com/containernetworking/plugins/blob/master/pkg/testutils/netns_linux.go
// notes: DO NOT open the nsPath like above repo do, or pouchd will hold the reference of the created network namespace.
//        pouchd will fail to remove the netns when stop pod sandbox.
func createNS(nsPath string) (res string, err error) {
	if err = os.MkdirAll(nsRunDir, 0755); err != nil {
		return "", err
	}

	// if the ns path is not given, create an empty file
	if nsPath == "" {
		b := make([]byte, 16)
		if _, err := rand.Reader.Read(b); err != nil {
			return "", errors.Wrap(err, "failed to generate random netns name")
		}

		nsName := fmt.Sprintf("cni-%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
		nsPath = path.Join(nsRunDir, nsName)
	}

	if _, err := os.Stat(nsPath); err != nil {
		if os.IsNotExist(err) {
			mountPointFd, err := os.Create(nsPath)
			if err != nil {
				return "", err
			}
			mountPointFd.Close()
		}
	}

	// Ensure the mount point is cleaned up on errors
	defer func() {
		if err != nil {
			os.RemoveAll(nsPath)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	// do namespace work in a dedicated goroutine, so that we can safely
	// Lock/Unlock OSThread without upsetting the lock/unlock state of
	// the caller of this function
	go (func() {
		defer wg.Done()
		runtime.LockOSThread()

		var origNS ns.NetNS
		origNS, err = ns.GetNS(getCurrentThreadNetNSPath())
		if err != nil {
			return
		}
		defer origNS.Close()

		// create a new netns on the current thread
		err = unix.Unshare(unix.CLONE_NEWNET)
		if err != nil {
			return
		}
		defer origNS.Set()

		// bind mount the new netns from the current thread onto the mount point
		err = unix.Mount(getCurrentThreadNetNSPath(), nsPath, "none", unix.MS_BIND, "")
		if err != nil {
			return
		}
	})()
	wg.Wait()

	if err != nil {
		unix.Unmount(nsPath, unix.MNT_DETACH)
		return "", errors.Wrapf(err, "failed to create namespace %s", nsPath)
	}

	return nsPath, nil
}
