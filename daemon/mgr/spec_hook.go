package mgr

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/storage/quota"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

//setup hooks specified by user via plugins, if set rich mode and init-script exists set init-script
func setupHook(ctx context.Context, c *Container, specWrapper *SpecWrapper) error {
	s := specWrapper.s
	if s.Hooks == nil {
		s.Hooks = &specs.Hooks{
			Prestart:  []specs.Hook{},
			Poststart: []specs.Hook{},
			Poststop:  []specs.Hook{},
		}
	}

	// setup plugin hook, if no hook plugin setup, skip this part.
	argsArr := specWrapper.argsArr
	prioArr := specWrapper.prioArr
	if len(argsArr) > 0 {
		var hookArr []*wrapperEmbedPrestart
		for i, hook := range s.Hooks.Prestart {
			hookArr = append(hookArr, &wrapperEmbedPrestart{-i, append([]string{hook.Path}, hook.Args...)})
		}
		for i, p := range prioArr {
			hookArr = append(hookArr, &wrapperEmbedPrestart{p, argsArr[i]})
		}
		sortedArr := hookArray(hookArr)
		sort.Sort(sortedArr)
		s.Hooks.Prestart = append(s.Hooks.Prestart, sortedArr.toOciPrestartHook()...)
	}

	// setup rich mode container hoopk, if no init script specified and no hook plugin setup, skip this part.
	if c.Config.Rich && c.Config.InitScript != "" {
		args := strings.Fields(c.Config.InitScript)
		if len(args) > 0 {
			s.Hooks.Prestart = append(s.Hooks.Prestart, specs.Hook{
				Path: args[0],
				Args: args[1:],
			})
		}
	}

	// setup diskquota hook, if rootFSQuota not set skip this part.
	if err := setRootfsDiskQuota(ctx, c, specWrapper); err != nil {
		return errors.Wrap(err, "failed to set rootfs disk")
	}

	// set volume mount tab
	if err := setMountTab(ctx, c, specWrapper); err != nil {
		return errors.Wrap(err, "failed to set volume mount tab prestart hook")
	}

	return nil
}

func setRootfsDiskQuota(ctx context.Context, c *Container, spec *SpecWrapper) error {
	rootFSQuota := quota.GetDefaultQuota(c.Config.DiskQuota)
	if rootFSQuota == "" {
		return nil
	}

	qid := "0"
	if c.Config.QuotaID != "" {
		qid = c.Config.QuotaID
	}

	target, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(os.Getpid()), "exe"))
	if err != nil {
		return err
	}

	spec.s.Hooks.Prestart = append(spec.s.Hooks.Prestart, specs.Hook{
		Path: target,
		Args: []string{"set-diskquota", c.BaseFS, rootFSQuota, qid},
	})

	return nil
}

func setMountTab(ctx context.Context, c *Container, spec *SpecWrapper) error {
	if len(c.BaseFS) == 0 {
		return nil
	}

	// set rootfs mount tab
	context := "/ / ext4 rw 0 0\n"
	if rootID, e := quota.GetDevID(c.BaseFS); e == nil {
		_, _, rootFsType := quota.CheckMountpoint(rootID)
		if len(rootFsType) > 0 {
			context = fmt.Sprintf("/ / %s rw 0 0\n", rootFsType)
		}
	}

	// set mount point tab
	i := 1
	for _, m := range c.Mounts {
		if m.Source == "" || m.Destination == "" {
			continue
		}

		finfo, err := os.Stat(m.Source)
		if err != nil || !finfo.IsDir() {
			continue
		}

		tempLine := fmt.Sprintf("/dev/v%02dd %s ext4 rw 0 0\n", i, m.Destination)
		if tmpID, e := quota.GetDevID(m.Source); e == nil {
			_, _, fsType := quota.CheckMountpoint(tmpID)
			if len(fsType) > 0 {
				tempLine = fmt.Sprintf("/dev/v%02dd %s %s rw 0 0\n", i, m.Destination, fsType)
			}
		}

		i++
		context += tempLine
	}

	// set shm mount tab
	context += "shm /dev/shm tmpfs rw 0 0\n"

	// save into mtab file.
	mtabPath := filepath.Join(c.BaseFS, "etc/mtab")
	hostmtabPath := filepath.Join(spec.ctrMgr.(*ContainerManager).Store.BaseDir, c.ID, "mtab")

	os.Remove(hostmtabPath)
	os.MkdirAll(filepath.Dir(hostmtabPath), 0755)
	err := ioutil.WriteFile(hostmtabPath, []byte(context), 0644)
	if err != nil {
		return fmt.Errorf("write %s failure", hostmtabPath)
	}

	mtabPrestart := specs.Hook{
		Path: "/bin/cp",
		Args: []string{"-f", "--remove-destination", hostmtabPath, mtabPath},
	}
	spec.s.Hooks.Prestart = append(spec.s.Hooks.Prestart, mtabPrestart)

	return nil
}

type hookArray []*wrapperEmbedPrestart

// Len is defined in order to support sort
func (h hookArray) Len() int {
	return len(h)
}

// Len is defined in order to support sort
func (h hookArray) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Less is defined in order to support sort, bigger priority execute first
func (h hookArray) Less(i, j int) bool {
	return h[i].Priority()-h[j].Priority() > 0
}

func (h hookArray) toOciPrestartHook() []specs.Hook {
	allHooks := make([]specs.Hook, len(h))
	for i, hook := range h {
		allHooks[i].Path = hook.Hook()[0]
		allHooks[i].Args = hook.Hook()[1:]
	}
	return allHooks
}

type wrapperEmbedPrestart struct {
	p    int
	args []string
}

func (w *wrapperEmbedPrestart) Priority() int {
	return w.p
}

func (w *wrapperEmbedPrestart) Hook() []string {
	return w.args
}
