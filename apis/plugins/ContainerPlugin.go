package plugins

// ContainerPlugin defines in which place a plugin will be triggered in container lifecycle
type ContainerPlugin interface {
	// PreCreate defines plugin point where recevives an container create request, in this plugin point user
	// could change the container create body passed-in by http request body
	PreCreate([]byte) ([]byte, error)

	// PreStart returns an array of PreStartHook which will pass to runc, in PreStartHook there is a Priority which
	// used to sort the pre start array that pass to runc, network plugin hook has priority value 0.
	PreStart([]byte)([]PreStartHook, error)
}

// PreStartHook defines the hook wrapper which will return priority and the hook info
type PreStartHook interface {
	// Priority returns priority of this hook, the bigger one will run first
	Priority() int
	// Hook returns the real hook command
	Hook() []string
}