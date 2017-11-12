package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/fatih/structs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Option uses to define the global options.
type Option struct {
	host    string
	timeout time.Duration
	TLS     utils.TLSConfig
}

// Cli is the client's core struct, it will be used to manage all subcommand, send http request
// to server and so on. it's a framework.
type Cli struct {
	Option
	rootCmd   *cobra.Command
	APIClient *client.APIClient
	padding   int
}

// NewCli creates an instance of 'Cli'.
func NewCli() *Cli {
	return &Cli{
		rootCmd: &cobra.Command{
			Use: "pouch",
		},
		padding: 3,
	}
}

// SetFlags sets all global options.
func (c *Cli) SetFlags() *Cli {
	cmd := c.rootCmd
	cmd.PersistentFlags().StringVarP(&c.Option.host, "host", "H", "unix:///var/run/pouchd.sock", "Specify listen address of pouchd")
	cmd.PersistentFlags().DurationVar(&c.Option.timeout, "timeout", time.Second*10, "Set timeout")
	utils.SetupTLSFlag(cmd.PersistentFlags(), &c.Option.TLS)
	return c
}

// NewAPIClient initializes the API client in Cli.
func (c *Cli) NewAPIClient() {
	client, err := client.NewAPIClient(c.Option.host, c.Option.TLS)
	if err != nil {
		logrus.Fatal(err)
	}
	c.APIClient = client
}

// Client returns API client torwards daemon.
func (c *Cli) Client() *client.APIClient {
	return c.APIClient
}

// Run executes the client program.
func (c *Cli) Run() {
	c.rootCmd.Execute()
}

// AddCommand add a subcommand.
func (c *Cli) AddCommand(parent, command Command) {
	command.Init(c)

	cmd := command.Cmd()

	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		c.NewAPIClient()
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		command.Run(args)
	}

	parent.Cmd().AddCommand(cmd)
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
			continue
		}

		display.AddRow(line)
	}

	display.Flush()
}

// NewRequest creates a HTTPRequest instance.
func (c *Cli) NewRequest(method, path string, obj interface{}, enc ...func(interface{}) (io.Reader, error)) (*Request, error) {
	serialize := func(o interface{}) (io.Reader, error) {
		if o == nil {
			return nil, nil
		}
		b, err := json.Marshal(o)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(b), nil
	}
	if len(enc) != 0 {
		serialize = enc[0]
	}

	body, err := serialize(obj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encoding object")
	}

	r, err := http.NewRequest(method, c.APIClient.BaseURL()+path, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to new http request")
	}
	return &Request{req: r, cli: c.APIClient.HTTPCli}, nil
}

// NewGetRequest creates an HTTP get request.
func (c *Cli) NewGetRequest(path string) (*Request, error) {
	return c.NewRequest(http.MethodGet, path, nil)
}

// NewPostRequest creates an HTTP post request.
func (c *Cli) NewPostRequest(path string, obj interface{}, enc ...func(interface{}) (io.Reader, error)) (*Request, error) {
	return c.NewRequest(http.MethodPost, path, obj, enc...)
}

// NewDeleteRequest creates an HTTP delete request.
func (c *Cli) NewDeleteRequest(path string, obj interface{}, enc ...func(interface{}) (io.Reader, error)) (*Request, error) {
	return c.NewRequest(http.MethodDelete, path, obj, enc...)
}
