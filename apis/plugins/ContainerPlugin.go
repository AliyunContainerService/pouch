package plugins

import "io"

// ContainerPlugin defines in which place a plugin will be triggered in container lifecycle
type ContainerPlugin interface {
	// PreCreate defines plugin point where recevives an container create request, in this plugin point user
	// could change the container create body passed-in by http request body
	PreCreate(io.ReadCloser) (io.ReadCloser, error)

	// PreStart returns an array of priority and args which will pass to runc, the every priority
	// used to sort the pre start array that pass to runc, network plugin hook always has priority value 0.
	PreStart(interface{}) ([]int, [][]string, error)
}
