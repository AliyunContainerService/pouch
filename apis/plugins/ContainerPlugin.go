package plugins

import "io"

// ContainerPlugin defines places where a plugin will be triggered in container lifecycle
type ContainerPlugin interface {
	// PreCreate defines plugin point where receives a container create request, in this plugin point user
	// could change the container create body passed-in by http request body
	PreCreate(io.ReadCloser) (io.ReadCloser, error)

	// PreStart returns an array of priority and args which will pass to runc, the every priority
	// used to sort the pre start array that pass to runc, network plugin hook always has priority value 0.
	PreStart(interface{}) ([]int, [][]string, error)

	//NetworkGenericParams accepts the container id and env of this container and returns the priority of this endpoint
	// and if this endpoint should enable resolver and a map which will be used as generic params to create endpoints of
	// this container
	PreCreateEndpoint(string, []string) (priority int, disableResolver bool, genericParam map[string]interface{})

	// PreUpdate defines plugin point where receives a container update request, in this plugin point user
	// could change the container update body passed-in by http request body
	PreUpdate(io.ReadCloser) (io.ReadCloser, error)
}
