package ctrd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/ioutils"

	"github.com/containerd/containerd"
	containerdtypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/archive"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/linux/runctypes"
	"github.com/containerd/containerd/oci"
	imagespec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	runtimeRoot = "/run"
)

type containerPack struct {
	id        string
	ch        chan *Message
	sch       <-chan containerd.ExitStatus
	container containerd.Container
	task      containerd.Task

	// client is to record which stream client the container connect with
	client        *WrapperClient
	skipStopHooks bool
}

// ContainerStats returns stats of the container.
func (c *Client) ContainerStats(ctx context.Context, id string) (*containerdtypes.Metric, error) {
	metric, err := c.containerStats(ctx, id)
	if err != nil {
		return metric, convertCtrdErr(err)
	}
	return metric, nil
}

// containerStats returns stats of the container.
func (c *Client) containerStats(ctx context.Context, id string) (*containerdtypes.Metric, error) {
	if !c.lock.TrylockWithRetry(ctx, id) {
		return nil, errtypes.ErrLockfailed
	}
	defer c.lock.Unlock(id)

	pack, err := c.watch.get(id)
	if err != nil {
		return nil, err
	}

	metrics, err := pack.task.Metrics(ctx)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// ExecContainer executes a process in container.
func (c *Client) ExecContainer(ctx context.Context, process *Process) error {
	if err := c.execContainer(ctx, process); err != nil {
		return convertCtrdErr(err)
	}
	return nil
}

// execContainer executes a process in container.
func (c *Client) execContainer(ctx context.Context, process *Process) error {
	pack, err := c.watch.get(process.ContainerID)
	if err != nil {
		return err
	}

	closeStdinCh := make(chan struct{})

	// make sure the closeStdinCh has been closed.
	defer func() {
		close(closeStdinCh)
	}()

	var (
		cntrID, execID          = pack.container.ID(), process.ExecID
		withStdin, withTerminal = process.IO.Stream().Stdin() != nil, process.P.Terminal
	)

	// create exec process in container
	execProcess, err := pack.task.Exec(ctx, process.ExecID, process.P, func(_ string) (cio.IO, error) {
		logrus.WithFields(
			logrus.Fields{
				"container": cntrID,
				"process":   execID,
			},
		).Debugf("creating cio (withStdin=%v, withTerminal=%v)", withStdin, withTerminal)

		fifoset, err := containerio.NewCioFIFOSet(execID, withStdin, withTerminal)
		if err != nil {
			return nil, err
		}
		return c.createIO(fifoset, cntrID, execID, closeStdinCh, process.IO.InitContainerIO)
	})
	if err != nil {
		return errors.Wrap(err, "failed to exec process")
	}

	// wait exec process to exit
	exitStatus, err := execProcess.Wait(context.TODO())
	if err != nil {
		return errors.Wrap(err, "failed to exec process")
	}

	errCh := make(chan error, 1)
	defer close(errCh)

	go func() {
		var msg *Message

		if startErr := <-errCh; startErr != nil {
			msg = &Message{
				err:      startErr,
				exitCode: 126,
				exitTime: time.Now().UTC(),
			}
		}

		// success to start which means the cmd is valid and wait
		// for process.
		if msg == nil {
			status := <-exitStatus
			msg = &Message{
				err:      status.Error(),
				exitCode: status.ExitCode(),
				exitTime: status.ExitTime(),
			}
		}

		// XXX: if exec process get run, io should be closed in this function,
		for _, hook := range c.hooks {
			if err := hook(process.ExecID, msg); err != nil {
				logrus.Errorf("failed to execute the exec exit hooks: %v", err)
				break
			}
		}

		// delete the finished exec process in containerd
		if _, err := execProcess.Delete(context.TODO()); err != nil {
			logrus.Warnf("failed to delete exec process %s: %s", process.ExecID, err)
		}
	}()

	// start the exec process
	if err := execProcess.Start(ctx); err != nil {
		errCh <- err
		return err
	}
	return nil
}

// ResizeExec changes the size of the TTY of the exec process running
// in the container to the given height and width.
func (c *Client) ResizeExec(ctx context.Context, id string, execid string, opts types.ResizeOptions) error {
	pack, err := c.watch.get(id)
	if err != nil {
		return err
	}

	execProcess, err := pack.task.LoadProcess(ctx, execid, nil)
	if err != nil {
		return err
	}

	return execProcess.Resize(ctx, uint32(opts.Width), uint32(opts.Height))
}

// ContainerPID returns the container's init process id.
func (c *Client) ContainerPID(ctx context.Context, id string) (int, error) {
	pid, err := c.containerPID(ctx, id)
	if err != nil {
		return pid, convertCtrdErr(err)
	}
	return pid, nil
}

// containerPID returns the container's init process id.
func (c *Client) containerPID(ctx context.Context, id string) (int, error) {
	pack, err := c.watch.get(id)
	if err != nil {
		return -1, err
	}
	return int(pack.task.Pid()), nil
}

// ContainerPIDs returns the all processes's ids inside the container.
func (c *Client) ContainerPIDs(ctx context.Context, id string) ([]int, error) {
	pids, err := c.containerPIDs(ctx, id)
	if err != nil {
		return pids, convertCtrdErr(err)
	}
	return pids, nil
}

// containerPIDs returns the all processes's ids inside the container.
func (c *Client) containerPIDs(ctx context.Context, id string) ([]int, error) {
	if !c.lock.TrylockWithRetry(ctx, id) {
		return nil, errtypes.ErrLockfailed
	}
	defer c.lock.Unlock(id)

	pack, err := c.watch.get(id)
	if err != nil {
		return nil, err
	}

	processes, err := pack.task.Pids(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get task's pids")
	}

	// convert []uint32 to []int.
	list := make([]int, 0, len(processes))
	for _, ps := range processes {
		list = append(list, int(ps.Pid))
	}
	return list, nil
}

// ProbeContainer probe the container's status, if timeout <= 0, will block to receive message.
func (c *Client) ProbeContainer(ctx context.Context, id string, timeout time.Duration) *Message {
	ch := c.watch.notify(id)

	if timeout <= 0 {
		msg := <-ch
		ch <- msg // put it back, make sure the method can be called repeatedly.

		return msg
	}
	select {
	case msg := <-ch:
		ch <- msg // put it back, make sure the method can be called repeatedly.
		return msg
	case <-time.After(timeout):
		return &Message{err: errtypes.ErrTimeout}
	case <-ctx.Done():
		return &Message{err: ctx.Err()}
	}
}

// RecoverContainer reload the container from metadata and watch it, if program be restarted.
func (c *Client) RecoverContainer(ctx context.Context, id string, io *containerio.IO) error {
	if err := c.recoverContainer(ctx, id, io); err != nil {
		return convertCtrdErr(err)
	}
	return nil
}

// recoverContainer reload the container from metadata and watch it, if program be restarted.
func (c *Client) recoverContainer(ctx context.Context, id string, io *containerio.IO) (err0 error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	if !c.lock.TrylockWithRetry(ctx, id) {
		return errtypes.ErrLockfailed
	}
	defer c.lock.Unlock(id)

	lc, err := wrapperCli.client.LoadContainer(ctx, id)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return errors.Wrapf(errtypes.ErrNotfound, "container %s", id)
		}
		return errors.Wrapf(err, "failed to load container(%s)", id)
	}

	task, err := lc.Task(ctx, func(fset *cio.FIFOSet) (cio.IO, error) {
		return c.attachIO(fset, io.InitContainerIO)
	})
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return errors.Wrap(err, "failed to get task")
		}
		// not found task, delete container directly.
		lc.Delete(ctx)
		return errors.Wrap(errtypes.ErrNotfound, "task")
	}

	statusCh, err := task.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to wait task")
	}

	c.watch.add(&containerPack{
		id:        id,
		container: lc,
		task:      task,
		ch:        make(chan *Message, 1),
		client:    wrapperCli,
		sch:       statusCh,
	})

	logrus.Infof("success to recover container: %s", id)
	return nil
}

// DestroyContainer kill container and delete it.
func (c *Client) DestroyContainer(ctx context.Context, id string, timeout int64) (*Message, error) {
	msg, err := c.destroyContainer(ctx, id, timeout)
	if err != nil {
		return msg, convertCtrdErr(err)
	}
	return msg, nil
}

// DestroyContainer kill container and delete it.
func (c *Client) destroyContainer(ctx context.Context, id string, timeout int64) (*Message, error) {
	// TODO(ziren): if we just want to stop a container,
	// we may need lease to lock the snapshot of container,
	// in case, it be deleted by gc.
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	ctx = leases.WithLease(ctx, wrapperCli.lease.ID())

	if !c.lock.TrylockWithRetry(ctx, id) {
		return nil, errtypes.ErrLockfailed
	}
	defer c.lock.Unlock(id)

	pack, err := c.watch.get(id)
	if err != nil {
		return nil, err
	}

	// if you call DestroyContainer to stop a container, will skip the hooks.
	// the caller need to execute the all hooks.
	pack.skipStopHooks = true
	defer func() {
		pack.skipStopHooks = false
	}()

	waitExit := func() *Message {
		return c.ProbeContainer(ctx, id, time.Duration(timeout)*time.Second)
	}

	var msg *Message

	// TODO: set task request timeout by context timeout
	if err := pack.task.Kill(ctx, syscall.SIGTERM, containerd.WithKillAll); err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, errors.Wrap(err, "failed to kill task")
		}
		goto clean
	}
	// wait for the task to exit.
	msg = waitExit()

	if err := msg.RawError(); err != nil && errtypes.IsTimeout(err) {
		// timeout, use SIGKILL to retry.
		if err := pack.task.Kill(ctx, syscall.SIGKILL, containerd.WithKillAll); err != nil {
			if !errdefs.IsNotFound(err) {
				return nil, errors.Wrap(err, "failed to kill task")
			}
			goto clean
		}
		msg = waitExit()
	}

	// ignore the error is stop time out
	// TODO: how to design the stop error is time out?
	if err := msg.RawError(); err != nil && !errtypes.IsTimeout(err) {
		return nil, err
	}

clean:
	if err := pack.container.Delete(ctx); err != nil {
		if !errdefs.IsNotFound(err) {
			return msg, errors.Wrap(err, "failed to delete container")
		}
	}

	logrus.Infof("success to destroy container: %s", id)

	return msg, c.watch.remove(ctx, id)
}

// PauseContainer pauses container.
func (c *Client) PauseContainer(ctx context.Context, id string) error {
	if err := c.pauseContainer(ctx, id); err != nil {
		return convertCtrdErr(err)
	}
	return nil
}

// pauseContainer pause container.
func (c *Client) pauseContainer(ctx context.Context, id string) error {
	if !c.lock.TrylockWithRetry(ctx, id) {
		return errtypes.ErrLockfailed
	}
	defer c.lock.Unlock(id)

	pack, err := c.watch.get(id)
	if err != nil {
		return err
	}

	if err := pack.task.Pause(ctx); err != nil {
		if !errdefs.IsNotFound(err) {
			return errors.Wrap(err, "failed to pause task")
		}
	}

	logrus.Infof("success to pause container: %s", id)

	return nil
}

// UnpauseContainer unpauses container.
func (c *Client) UnpauseContainer(ctx context.Context, id string) error {
	if err := c.unpauseContainer(ctx, id); err != nil {
		return convertCtrdErr(err)
	}
	return nil
}

// unpauseContainer unpauses a container.
func (c *Client) unpauseContainer(ctx context.Context, id string) error {
	if !c.lock.TrylockWithRetry(ctx, id) {
		return errtypes.ErrLockfailed
	}
	defer c.lock.Unlock(id)

	pack, err := c.watch.get(id)
	if err != nil {
		return err
	}

	if err := pack.task.Resume(ctx); err != nil {
		if !errdefs.IsNotFound(err) {
			return errors.Wrap(err, "failed to resume task")
		}
	}

	logrus.Infof("success to unpause container: %s", id)

	return nil
}

// CreateContainer create container and start process.
func (c *Client) CreateContainer(ctx context.Context, container *Container, checkpointDir string) error {
	var (
		ref = container.Image
		id  = container.ID
	)

	if !c.lock.TrylockWithRetry(ctx, id) {
		return errtypes.ErrLockfailed
	}
	defer c.lock.Unlock(id)

	if err := c.createContainer(ctx, ref, id, checkpointDir, container); err != nil {
		return convertCtrdErr(err)
	}
	return nil
}

func (c *Client) createContainer(ctx context.Context, ref, id, checkpointDir string, container *Container) (err0 error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	// if creating the container by specify rootfs, we no need use the image
	if !container.RootFSProvided {
		// get image
		img, err := wrapperCli.client.GetImage(ctx, ref)
		if err != nil {
			if errdefs.IsNotFound(err) {
				return errors.Wrapf(errtypes.ErrNotfound, "image %s", ref)
			}
			return errors.Wrapf(err, "failed to get image %s", ref)
		}

		logrus.Infof("success to get image %s, container id %s", img.Name(), id)
	}

	// create container
	options := []containerd.NewContainerOpts{
		containerd.WithSnapshotter(CurrentSnapshotterName(ctx)),
		containerd.WithContainerLabels(container.Labels),
		containerd.WithRuntime(fmt.Sprintf("io.containerd.runtime.v1.%s", runtime.GOOS), &runctypes.RuncOptions{
			Runtime:     container.Runtime,
			RuntimeRoot: runtimeRoot,
		}),
	}

	rootFSPath := "rootfs"
	// if container is taken over by pouch, not created by pouch
	if container.RootFSProvided {
		rootFSPath = container.BaseFS
	} else { // containers created by pouch must first create snapshot
		// check snapshot exist or not.
		if _, err := c.GetSnapshot(ctx, container.SnapshotID); err != nil {
			return errors.Wrapf(err, "failed to create container %s", id)
		}
		options = append(options, containerd.WithSnapshot(container.SnapshotID))
	}

	// specify Spec for new container
	specOptions := []oci.SpecOpts{
		oci.WithRootFSPath(rootFSPath),
	}
	options = append(options, containerd.WithSpec(container.Spec, specOptions...))

	nc, err := wrapperCli.client.NewContainer(ctx, id, options...)
	if err != nil {
		return errors.Wrapf(err, "failed to create container %s", id)
	}

	defer func() {
		if err0 != nil {
			// Delete snapshot when start failed, may cause data lost.
			nc.Delete(ctx)
		}
	}()

	logrus.Infof("success to new container: %s", id)

	// create task
	pack, err := c.createTask(ctx, id, checkpointDir, nc, container, wrapperCli.client)
	if err != nil {
		return err
	}

	// add grpc client to pack struct
	pack.client = wrapperCli

	c.watch.add(pack)

	return nil
}

func (c *Client) createTask(ctx context.Context, id, checkpointDir string, container containerd.Container, cc *Container, client *containerd.Client) (p *containerPack, err0 error) {
	var pack *containerPack

	checkpoint, err := createCheckpointDescriptor(ctx, checkpointDir, client)
	if err != nil {
		return pack, errors.Wrapf(err, "failed to create checkpoint descriptor")
	}
	defer func() {
		if checkpoint != nil {
			// remove the checkpoint blob after task start
			err := client.ContentStore().Delete(context.Background(), checkpoint.Digest)
			if err != nil {
				logrus.Warnf("failed to delete temporary checkpoint entry: %s", err)
			}
		}
	}()

	var (
		cntrID, execID          = id, id
		withStdin, withTerminal = cc.IO.Stream().Stdin() != nil, cc.Spec.Process.Terminal
		closeStdinCh            = make(chan struct{})
	)

	// create task
	task, err := container.NewTask(ctx, func(_ string) (cio.IO, error) {
		logrus.WithField("container", cntrID).Debugf("creating cio (withStdin=%v, withTerminal=%v)", withStdin, withTerminal)

		fifoset, err := containerio.NewCioFIFOSet(execID, withStdin, withTerminal)
		if err != nil {
			return nil, err
		}
		return c.createIO(fifoset, cntrID, execID, closeStdinCh, cc.IO.InitContainerIO)
	}, withCheckpointOpt(checkpoint))
	close(closeStdinCh)

	if err != nil {
		return pack, errors.Wrapf(err, "failed to create task for container(%s)", id)
	}

	defer func() {
		if err0 != nil {
			task.Delete(ctx)
		}
	}()

	statusCh, err := task.Wait(context.TODO())
	if err != nil {
		return pack, errors.Wrapf(err, "failed to wait task in container(%s)", id)
	}

	logrus.Infof("success to create task(pid=%d) in container(%s)", task.Pid(), id)

	// start task
	if err := task.Start(ctx); err != nil {
		return pack, errors.Wrapf(err, "failed to start task(%d) in container(%s)", task.Pid(), id)
	}

	logrus.Infof("success to start task in container(%s)", id)

	pack = &containerPack{
		id:        id,
		container: container,
		task:      task,
		ch:        make(chan *Message, 1),
		sch:       statusCh,
	}

	return pack, nil
}

// UpdateResources updates the configurations of a container.
func (c *Client) UpdateResources(ctx context.Context, id string, resources types.Resources) error {
	if err := c.updateResources(ctx, id, resources); err != nil {
		return convertCtrdErr(err)
	}
	return nil
}

// updateResources updates the configurations of a container.
func (c *Client) updateResources(ctx context.Context, id string, resources types.Resources) error {
	if !c.lock.TrylockWithRetry(ctx, id) {
		return errtypes.ErrLockfailed
	}
	defer c.lock.Unlock(id)

	pack, err := c.watch.get(id)
	if err != nil {
		return err
	}

	r, err := toLinuxResources(resources)
	if err != nil {
		return err
	}

	return pack.task.Update(ctx, containerd.WithResources(r))
}

// ResizeContainer changes the size of the TTY of the init process running
// in the container to the given height and width.
func (c *Client) ResizeContainer(ctx context.Context, id string, opts types.ResizeOptions) error {
	if err := c.resizeContainer(ctx, id, opts); err != nil {
		return convertCtrdErr(err)
	}
	return nil
}

// resizeContainer changes the size of the TTY of the init process running
// in the container to the given height and width.
func (c *Client) resizeContainer(ctx context.Context, id string, opts types.ResizeOptions) error {
	if !c.lock.TrylockWithRetry(ctx, id) {
		return errtypes.ErrLockfailed
	}
	defer c.lock.Unlock(id)

	pack, err := c.watch.get(id)
	if err != nil {
		return err
	}

	return pack.task.Resize(ctx, uint32(opts.Width), uint32(opts.Height))
}

// WaitContainer waits until container's status is stopped.
func (c *Client) WaitContainer(ctx context.Context, id string) (types.ContainerWaitOKBody, error) {
	waitBody, err := c.waitContainer(ctx, id)
	if err != nil {
		return waitBody, convertCtrdErr(err)
	}
	return waitBody, nil
}

// waitContainer waits until container's status is stopped.
func (c *Client) waitContainer(ctx context.Context, id string) (types.ContainerWaitOKBody, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return types.ContainerWaitOKBody{}, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	ctx = leases.WithLease(ctx, wrapperCli.lease.ID())

	waitExit := func() *Message {
		return c.ProbeContainer(ctx, id, -1*time.Second)
	}

	// wait for the task to exit.
	msg := waitExit()

	errMsg := ""
	err = msg.RawError()
	if err != nil {
		if errtypes.IsTimeout(err) {
			return types.ContainerWaitOKBody{}, err
		}
		errMsg = err.Error()
	}

	return types.ContainerWaitOKBody{
		Error:      errMsg,
		StatusCode: int64(msg.ExitCode()),
	}, nil
}

// CreateCheckpoint create a checkpoint from a running container
func (c *Client) CreateCheckpoint(ctx context.Context, id string, checkpointDir string, exit bool) error {
	pack, err := c.watch.get(id)
	if err != nil {
		return err
	}

	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}
	client := wrapperCli.client

	var opts []containerd.CheckpointTaskOpts
	if exit {
		opts = append(opts, containerd.WithExit)
	}
	checkpoint, err := pack.task.Checkpoint(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to checkpoint: %s", err)
	}
	// delete image since it is a checkpoint-format image, can not
	// distinguished when load images.
	defer client.ImageService().Delete(ctx, checkpoint.Name())

	return applyCheckpointImage(ctx, client, checkpoint, checkpointDir)
}

func applyCheckpointImage(ctx context.Context, client *containerd.Client, checkpoint containerd.Image, checkpointDir string) error {
	b, err := content.ReadBlob(ctx, client.ContentStore(), checkpoint.Target().Digest)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve checkpoint data")
	}
	var index imagespec.Index
	if err := json.Unmarshal(b, &index); err != nil {
		return errors.Wrapf(err, "failed to decode checkpoint data")
	}

	var cpDesc *imagespec.Descriptor
	for _, m := range index.Manifests {
		if m.MediaType == images.MediaTypeContainerd1Checkpoint {
			cpDesc = &m
			break
		}
	}
	if cpDesc == nil {
		return errors.Wrapf(err, "invalid checkpoint")
	}

	rat, err := client.ContentStore().ReaderAt(ctx, cpDesc.Digest)
	if err != nil {
		return errors.Wrapf(err, "failed to get checkpoint reader")
	}
	defer rat.Close()
	_, err = archive.Apply(ctx, checkpointDir, content.NewReader(rat))
	if err != nil {
		return errors.Wrapf(err, "failed to read checkpoint reader")
	}

	return nil
}

func writeContent(ctx context.Context, mediaType, ref string, r io.Reader, client *containerd.Client) (*containerdtypes.Descriptor, error) {
	writer, err := client.ContentStore().Writer(ctx, ref, 0, "")
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	size, err := io.Copy(writer, r)
	if err != nil {
		return nil, err
	}
	labels := map[string]string{
		"containerd.io/gc.root": time.Now().UTC().Format(time.RFC3339),
	}
	if err := writer.Commit(ctx, 0, "", content.WithLabels(labels)); err != nil {
		return nil, err
	}
	return &containerdtypes.Descriptor{
		MediaType: mediaType,
		Digest:    writer.Digest(),
		Size_:     size,
	}, nil

}

func createCheckpointDescriptor(ctx context.Context, checkpointDir string, client *containerd.Client) (*containerdtypes.Descriptor, error) {
	if checkpointDir == "" {
		return nil, nil
	}

	// create a checkpoint blob
	tar := archive.Diff(ctx, "", checkpointDir)
	checkpoint, err := writeContent(ctx, images.MediaTypeContainerd1Checkpoint, checkpointDir, tar, client)
	if err := tar.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close checkpoint tar stream")
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to upload checkpoint to containerd")
	}

	return checkpoint, nil
}

func withCheckpointOpt(checkpoint *containerdtypes.Descriptor) containerd.NewTaskOpts {
	return func(_ context.Context, _ *containerd.Client, t *containerd.TaskInfo) error {
		t.Checkpoint = checkpoint
		return nil
	}
}

// InitStdio allows caller to handle any initialize job.
type InitStdio func(dio *containerio.DirectIO) (cio.IO, error)

func (c *Client) createIO(fifoSet *containerio.CioFIFOSet, cntrID, procID string, closeStdinCh <-chan struct{}, initstdio InitStdio) (cio.IO, error) {
	cdio, err := containerio.NewDirectIO(context.Background(), fifoSet)
	if err != nil {
		return nil, err
	}

	if cdio.Stdin != nil {
		var (
			errClose  error
			stdinOnce sync.Once
		)
		oldStdin := cdio.Stdin
		cdio.Stdin = ioutils.NewWriteCloserWrapper(oldStdin, func() error {
			stdinOnce.Do(func() {
				errClose = oldStdin.Close()

				// Both the caller and container/exec process holds write side pipe
				// for the stdin. When the caller closes the write pipe, the process doesn't
				// exit until the caller calls the CloseIO.
				go func() {
					<-closeStdinCh
					if err := c.closeStdinIO(cntrID, procID); err != nil {
						// TODO(fuweid): for the CloseIO grpc call, the containerd doesn't
						// return correct status code if the process doesn't exist.
						// for the case, we should use strings.Contains to reduce warning
						// log. it will be fixed in containerd#2747.
						if !errdefs.IsNotFound(err) && !strings.Contains(err.Error(), "not found") {
							logrus.WithError(err).Warnf("failed to close stdin containerd IO (container:%v, process:%v", cntrID, procID)
						}
					}
				}()
			})
			return errClose
		})
	}

	cntrio, err := initstdio(cdio)
	if err != nil {
		cdio.Cancel()
		cdio.Close()
		return nil, err
	}
	return cntrio, nil
}

func (c *Client) attachIO(fifoSet *cio.FIFOSet, initstdio InitStdio) (cio.IO, error) {
	if fifoSet == nil {
		return nil, fmt.Errorf("cannot attach to existing fifos")
	}

	cdio, err := containerio.NewDirectIO(context.Background(), &containerio.CioFIFOSet{
		Config: cio.Config{
			Terminal: fifoSet.Terminal,
			Stdin:    fifoSet.In,
			Stdout:   fifoSet.Out,
			Stderr:   fifoSet.Err,
		},
	})
	if err != nil {
		return nil, err
	}

	cntrio, err := initstdio(cdio)
	if err != nil {
		cdio.Cancel()
		cdio.Close()
		return nil, err
	}
	return cntrio, nil
}

// closeStdinIO is used to close the write side of fifo in containerd-shim.
//
// NOTE: we should use client to make rpc call directly. if we retrieve it from
// watch, it might return 404 because the pack is saved into cache after Start.
func (c *Client) closeStdinIO(containerID, processID string) error {
	ctx := context.Background()
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	cli := wrapperCli.client
	cntr, err := cli.LoadContainer(ctx, containerID)
	if err != nil {
		return err
	}

	t, err := cntr.Task(ctx, nil)
	if err != nil {
		return err
	}

	p, err := t.LoadProcess(ctx, processID, nil)
	if err != nil {
		return err
	}

	return p.CloseIO(ctx, containerd.WithStdinCloser)
}
