package main

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/term"

	"github.com/spf13/cobra"
)

// loginDescription is used to describe login command and auto generate command doc.
var loginDescription = "\nlogin to a v1/v2 registry with the provided credentials."

// LoginCommand use to implement 'login' command.
type LoginCommand struct {
	baseCommand

	username string
	password string
}

// Init initialize login command.
func (l *LoginCommand) Init(c *Cli) {
	l.cli = c

	l.cmd = &cobra.Command{
		Use:   "login [OPTIONS] [SERVER]",
		Short: "Login to a registry",
		Long:  loginDescription,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return l.runLogin(args)
		},
		Example: loginExample(),
	}
	l.addFlags()
}

// addFlags adds flags for specific command.
func (l *LoginCommand) addFlags() {
	flagSet := l.cmd.Flags()

	flagSet.StringVarP(&l.username, "username", "u", "", "username for registry")
	flagSet.StringVarP(&l.password, "password", "p", "", "password for registry")
}

// runLogin is the entry of login command.
func (l *LoginCommand) runLogin(args []string) error {
	auth := configureAuth(l.username, l.password)
	if len(args) > 0 {
		auth.ServerAddress = args[0]
	}

	ctx := context.Background()
	apiClient := l.cli.Client()
	resp, err := apiClient.RegistryLogin(ctx, auth)
	if err != nil {
		return err
	}

	if resp.Status != "" {
		fmt.Printf("%s\n", resp.Status)
	}

	return nil
}

func configureAuth(username, password string) *types.AuthConfig {
	if username == "" {
		username = readInput("Username", true)
	}

	if password == "" {
		password = readInput("Password", false)
	}

	return &types.AuthConfig{
		Username: username,
		Password: password,
	}
}

func readInput(prompt string, echo bool) (read string) {
	fmt.Printf("%s: ", prompt)
	if echo {
		fmt.Scanf("%s", &read)
	} else {
		disableEcho(echo, &read)
	}
	return
}

func disableEcho(echo bool, read *string) {
	term.StdinEcho(false)
	fmt.Scanf("%s", read)
	fmt.Printf("\n")

	// restore stdin terminal
	term.StdinEcho(true)
}

// loginExample shows examples in login command, and is used in auto-generated cli docs.
func loginExample() string {
	return `$ pouch login -u $username -p $password
Login Succeeded`
}
