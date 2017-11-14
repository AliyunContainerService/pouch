package ctrd

import (
	"context"
	"syscall"
	"time"

	"github.com/alibaba/pouch/daemon/containerio"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type containerPack struct {
	id        string
	ch        chan *Message
	sch       <-chan containerd.ExitStatus
	container containerd.Container
	task      containerd.Task
}

// ContainerPID returns the container's init process id.
func (c *Client) ContainerPID(ctx context.Context, id string) (int, error) {
	pack, err := c.watch.get(id)
	if err != nil {
		return -1, err
	}
	return int(pack.task.Pid()), nil
}

// ContainerPIDs returns the all processes's ids inside the container.
func (c *Client) ContainerPIDs(ctx context.Context, id string) ([]int, error) {
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
		return &Message{err: ErrTimeout}
	case <-ctx.Done():
		return &Message{err: ctx.Err()}
	}
}

// RecoverContainer reload the container from metadata and watch it, if program be restarted.
func (c *Client) RecoverContainer(ctx context.Context, id string, io *containerio.IO) error {
	if !c.lock.Trylock(id) {
		return ErrTrylockFailed
	}
	defer c.lock.Unlock(id)

	lc, err := c.client.LoadContainer(ctx, id)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return ErrContainerNotfound
		}
		return errors.Wrap(err, "failed to load container")
	}

	task, err := lc.Task(ctx, containerd.WithAttach(io.Stdin, io.Stdout, io.Stderr))
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return errors.Wrap(err, "failed to get task")
		}
		// not found task, delete container directly.
		lc.Delete(ctx, containerd.WithSnapshotCleanup)
		return ErrTaskNotfound
	}

	statusCh, err := task.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to wait task")
	}
	c.watch.add(containerPack{
		id:        id,
		container: lc,
		task:      task,
		ch:        make(chan *Message, 1),
		sch:       statusCh,
	})

	logrus.Infof("success to recover container: %s", id)
	return nil
}

// DestroyContainer kill container and delete it.
func (c *Client) DestroyContainer(ctx context.Context, id string) (*Message, error) {
	if !c.lock.Trylock(id) {
		return nil, ErrTrylockFailed
	}
	defer c.lock.Unlock(id)

	pack, err := c.watch.get(id)
	if err != nil {
		return nil, err
	}

	waitExit := func() *Message {
		return c.ProbeContainer(ctx, id, time.Second*5)
	}

	var msg *Message

	if err := pack.task.Kill(ctx, syscall.SIGTERM, containerd.WithKillAll); err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, errors.Wrap(err, "failed to kill task")
		}
	} else {
		// wait for the task to exit.
		msg = waitExit()
	}

	if msg.Error() != nil {
		if cerr, ok := msg.Error().(Error); ok && cerr.IsTimeout() {
			// timeout, use SIGKILL to retry.
			if err := pack.task.Kill(ctx, syscall.SIGKILL, containerd.WithKillAll); err != nil {
				if !errdefs.IsNotFound(err) {
					return nil, errors.Wrap(err, "failed to kill task")
				}

			} else {
				msg = waitExit()
			}
		}
	}
	if msg.Error() != nil {
		if cerr, ok := msg.Error().(Error); ok && cerr.IsTimeout() {
			return nil, cerr
		}
	}

	if err := pack.container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		if !errdefs.IsNotFound(err) {
			return msg, errors.Wrap(err, "failed to delete container")
		}
	}

	logrus.Infof("success to destroy container: %s", id)

	return msg, c.watch.remove(ctx, id)
}

// CreateContainer create container and start process.
func (c *Client) CreateContainer(ctx context.Context, container *Container) error {
	var (
		ref = container.Info.Config.Image
		id  = container.Info.ID
	)

	if !c.lock.Trylock(id) {
		return ErrTrylockFailed
	}
	defer c.lock.Unlock(id)

	return c.createContainer(ctx, ref, id, container)
}

func (c *Client) createContainer(ctx context.Context, ref, id string, container *Container) (err0 error) {
	// get image
	img, err := c.client.GetImage(ctx, ref)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return ErrImageNotfound
		}
		return errors.Wrapf(err, "failed to get image: %s", ref)
	}

	logrus.Infof("success to get image: %s, container id: %s", img.Name(), id)

	// create container
	specOptions := []containerd.SpecOpts{
		containerd.WithImageConfig(img),
		containerd.WithRootFSPath("rootfs"),
	}
	if args := container.Spec.Process.Args; len(args) != 0 {
		specOptions = append(specOptions, containerd.WithProcessArgs(args...))
	}

	options := []containerd.NewContainerOpts{
		containerd.WithNewSnapshot(id, img),
	}
	if container.Spec != nil {
		options = append(options, containerd.WithSpec(container.Spec, specOptions...))
	} else {
		options = append(options, containerd.WithNewSpec(specOptions...))
	}

	nc, err := c.client.NewContainer(ctx, id, options...)
	if err != nil {
		return errors.Wrapf(err, "failed to create container, id: %s", id)
	}

	defer func() {
		if err0 != nil {
			nc.Delete(ctx, containerd.WithSnapshotCleanup)
		}
	}()

	logrus.Infof("success to new container: %s", id)

	// create task
	pack, err := c.createTask(ctx, id, nc, container)
	if err != nil {
		return err
	}

	c.watch.add(pack)

	return nil
}

func (c *Client) createTask(ctx context.Context, id string, container containerd.Container, cc *Container) (p containerPack, err0 error) {
	var pack containerPack

	var io containerd.IOCreation
	if cc.Spec.Process.Terminal {
		io = containerd.NewIOWithTerminal(cc.IO.Stdin, cc.IO.Stdout, cc.IO.Stderr, true)
	} else {
		io = containerd.NewIO(cc.IO.Stdin, cc.IO.Stdout, cc.IO.Stderr)
	}

	// create task
	task, err := container.NewTask(ctx, io)
	if err != nil {
		return pack, errors.Wrapf(err, "failed to create task, container id: %s", id)
	}

	defer func() {
		if err0 != nil {
			task.Delete(ctx)
		}
	}()

	statusCh, err := task.Wait(context.TODO())
	if err != nil {
		return pack, errors.Wrap(err, "failed to wait task")
	}

	logrus.Infof("success to new task, container id: %s, pid: %d", id, task.Pid())

	// start task
	if err := task.Start(ctx); err != nil {
		return pack, errors.Wrapf(err, "failed to start task: %d, container id: %s", task.Pid(), id)
	}

	logrus.Infof("success to start task, container id: %s", id)

	pack = containerPack{
		id:        id,
		container: container,
		task:      task,
		ch:        make(chan *Message, 1),
		sch:       statusCh,
	}

	return pack, nil
}

func (c *Client) listContainerStore(ctx context.Context) ([]string, error) {
	containers, err := c.client.ContainerService().List(ctx)
	if err != nil {
		return nil, err
	}

	var cs []string

	for _, c := range containers {
		cs = append(cs, c.ID)
	}

	return cs, nil
}
