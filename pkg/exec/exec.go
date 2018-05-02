package exec

import (
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Process represents a program will be execute.
type Process struct {
	// Binary's absolute path.
	Path string

	// Arguments use to execute binary.
	Args []string

	// Output represents a file that the stdio and stderr will write.
	Output string

	sync.Mutex

	running   bool
	pid       int
	cmd       *exec.Cmd
	file      *os.File
	forceStop bool
}

// Stop kill the running process.
func (p *Process) Stop() error {
	p.Lock()
	if !p.running {
		p.Unlock()
		return nil
	}

	// forceStop=true will not restart the process.
	p.forceStop = true

	if err := p.cmd.Process.Kill(); err != nil {
		return errors.Wrap(err, "kill process")
	}

	// unlock process
	p.Unlock()
	return nil
}

// Start start process.
func (p *Process) Start() error {
	// initialize command.
	cmd := exec.Command(p.Path, p.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var (
		file *os.File
		err  error
	)
	if p.Output != "" {
		if file, err = os.OpenFile(p.Output, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666); err != nil {
			return errors.Wrap(err, "failed to open file")
		}
		p.file = file
		cmd.Stdout = file
		cmd.Stderr = file
	}

	if err := cmd.Start(); err != nil {
		if p.file != nil {
			p.file.Close()
		}
		return err
	}

	p.running = true
	p.pid = cmd.Process.Pid
	p.cmd = cmd

	// use a goroutine to wait process.
	go func(p *Process) {
		if err := p.cmd.Wait(); err != nil {
			logrus.Errorf("failed to wait process: %s, %v", p.Path, err)
		}

		logrus.Warnf("process: %s exited", p.Path)

		p.Lock()
		defer p.Unlock()

		p.running = false
		p.pid = -1
		if p.file != nil {
			p.file.Close()
		}
	}(p)

	return nil
}

// Processes is slice of Process type.
type Processes []*Process

// StopAll stop the all running processes.
func (ps Processes) StopAll() error {
	for _, p := range ps {
		if err := p.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// RunAll execute the all commands.
func (ps Processes) RunAll() error {
	for _, p := range ps {
		if err := p.Start(); err != nil {
			return err
		}
	}

	// monitor all processes, restart them when exited.
	go func() {
		for {
			<-time.After(time.Second)

			for _, p := range ps {
				p.Lock()

				if !p.forceStop && !p.running {
					if err := p.Start(); err != nil {
						logrus.Errorf("failed to restart process: %v", err)
					}
				}

				p.Unlock()
			}
		}
	}()

	return nil
}
