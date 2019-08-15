package main

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/log"

	"github.com/fatih/structs"
	"github.com/spf13/cobra"
)

// pouchDescription is used to describe pouch command in detail and auto generate command doc.
var pouchDescription = "pouch is a client side tool pouch to interact with daemon side process pouchd. " +
	"Flags and arguments can be input to do what actually you wish. " +
	"Then pouch parses the flags and arguments and sends a RESTful request to daemon side pouchd."

// Option uses to define the global options.
type Option struct {
	host  string
	Debug bool
	TLS   client.TLSConfig
}

// Cli is the client's core struct, it will be used to manage all subcommand, send http request
// to server and so on. it's a framework.
type Cli struct {
	Option
	rootCmd   *cobra.Command
	APIClient client.CommonAPIClient
	padding   int
}

// NewCli creates an instance of 'Cli'.
func NewCli() *Cli {
	return &Cli{
		rootCmd: &cobra.Command{
			Use:   "pouch",
			Short: "An efficient container engine",
			Long:  pouchDescription,
			// disable displaying auto generation tag in cli docs
			DisableAutoGenTag: true,
		},
		padding: 3,
	}
}

// SetFlags sets all global options.
func (c *Cli) SetFlags() *Cli {
	flags := c.rootCmd.PersistentFlags()
	flags.StringVarP(&c.Option.host, "host", "H", "unix:///var/run/pouchd.sock", "Specify connecting address of Pouch CLI")
	flags.BoolVarP(&c.Option.Debug, "debug", "D", false, "Switch client log level to DEBUG mode")
	flags.StringVar(&c.Option.TLS.Key, "tlskey", "", "Specify key file of TLS")
	flags.StringVar(&c.Option.TLS.Cert, "tlscert", "", "Specify cert file of TLS")
	flags.StringVar(&c.Option.TLS.CA, "tlscacert", "", "Specify CA file of TLS")
	flags.BoolVar(&c.Option.TLS.VerifyRemote, "tlsverify", false, "Use TLS and verify remote")
	return c
}

// NewAPIClient initializes the API client in Cli.
func (c *Cli) NewAPIClient() {
	client, err := client.NewAPIClient(c.Option.host, c.Option.TLS)
	if err != nil {
		log.With(nil).Fatal(err)
	}

	c.APIClient = client
}

// Client returns API client towards daemon.
func (c *Cli) Client() client.CommonAPIClient {
	return c.APIClient
}

// Run executes the client program.
func (c *Cli) Run() error {
	return c.rootCmd.Execute()
}

// AddCommand add a subcommand.
func (c *Cli) AddCommand(parent, child Command) {
	child.Init(c)

	parentCmd := parent.Cmd()
	childCmd := child.Cmd()

	// make command error not return command usage and error
	childCmd.SilenceUsage = true
	childCmd.SilenceErrors = true
	childCmd.DisableFlagsInUseLine = true

	childCmd.PreRun = func(cmd *cobra.Command, args []string) {
		log.Init(c.Debug)
		c.NewAPIClient()
	}

	parentCmd.AddCommand(childCmd)
}

// NewTableDisplay creates a display instance, and uses to format output with table.
func (c *Cli) NewTableDisplay() *Display {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, c.padding, ' ', 0)
	return &Display{w}
}

// Print outputs the obj's fields.
func (c *Cli) Print(obj interface{}) {
	display := c.NewTableDisplay()
	kvs := structs.Map(obj)

	for k, v := range kvs {
		line := []string{k + ":"}

		switch v.(type) {
		case string:
			line = append(line, v.(string))

		case int:
			line = append(line, strconv.FormatInt(int64(v.(int)), 10))

		case int32:
			line = append(line, strconv.FormatInt(int64(v.(int32)), 10))

		case int64:
			line = append(line, strconv.FormatInt(v.(int64), 10))

		case bool:
			if v.(bool) {
				line = append(line, "true")
			} else {
				line = append(line, "false")
			}

		default:
			line = append(line, fmt.Sprintf("%v", v))
		}

		display.AddRow(line)
	}

	display.Flush()
}

// ExitError defines exit error produce by cli commands.
type ExitError struct {
	Code   int
	Status string
}

// Error implements error interface.
func (e ExitError) Error() string {
	return fmt.Sprintf("Exit Code: %d, Status: %s", e.Code, e.Status)
}
