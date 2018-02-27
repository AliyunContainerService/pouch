package main

import (
	"fmt"
	"os"

	"github.com/alibaba/pouch/credential"

	"github.com/spf13/cobra"
)

// logoutDescription is used to describe logout command and auto generate command doc.
var logoutDescription = "\nlogout from a v1/v2 registry."

// LogoutCommand use to implement 'logout' command.
type LogoutCommand struct {
	baseCommand
}

// Init initialize logout command.
func (l *LogoutCommand) Init(c *Cli) {
	l.cli = c

	l.cmd = &cobra.Command{
		Use:   "logout [SERVER]",
		Short: "Logout from a registry",
		Long:  logoutDescription,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return l.runLogout(args)
		},
		Example: logoutExample(),
	}
}

// runLogout is the entry of logout command.
func (l *LogoutCommand) runLogout(args []string) error {
	var serverAddress string
	if len(args) > 0 {
		serverAddress = args[0]
	}

	// TODO: better way to get registry server address, or get it from a default variable
	registry := serverAddress
	if registry == "" {
		registry = "https://index.docker.io"
	}

	if !credential.Exist(serverAddress) {
		fmt.Fprintf(os.Stdout, "Has not logged in registry: %s\n", registry)
		return nil
	}

	if err := credential.Delete(serverAddress); err != nil {
		fmt.Fprintf(os.Stderr, "Fail to remove login credential: %s\n", err)
		return err
	}

	fmt.Fprintf(os.Stdout, "Remove login credential for registry:%s\n", registry)
	return nil
}

// logoutExample shows examples in logout command, and is used in auto-generated cli docs.
func logoutExample() string {
	return `$ pouch logout $registry
Remove login credential for registry: $registry`
}
