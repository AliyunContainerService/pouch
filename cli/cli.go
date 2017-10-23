package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/structs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Option use to define the global options.
type Option struct {
	host    string
	timeout time.Duration
}

// Cli is the client's core struct, it will be used to manage all subcommand, send http request
// to server and so on. it's a framework.
type Cli struct {
	Option
	rootCmd   *cobra.Command
	transport *http.Transport
	httpcli   *http.Client
	padding   int
	baseURL   string
}

// NewCli create a instance of 'Cli'.
func NewCli() *Cli {
	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Second*10)
		},
	}

	return &Cli{
		rootCmd: &cobra.Command{
			Use: "pouch",
		},
		httpcli: &http.Client{
			Transport: tr,
			Timeout:   time.Second * 30,
		},
		transport: tr,
		padding:   3,
	}
}

// SetFlags set all global options.
func (c *Cli) SetFlags() *Cli {
	cmd := c.rootCmd
	cmd.PersistentFlags().StringVarP(&c.Option.host, "host", "H", "unix:///var/run/pouchd.sock", "Specify listen address of pouchd")
	cmd.PersistentFlags().DurationVar(&c.Option.timeout, "timeout", time.Second*10, "Set timeout")

	return c
}

// Run execute the client program.
func (c *Cli) Run() {
	c.rootCmd.Execute()
}

// AddCommand add a subcommand.
func (c *Cli) AddCommand(command SubCommand) {
	command.Init(c)

	cmd := command.Cmd()

	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		if strings.HasPrefix(c.host, "unix://") {
			dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.DialTimeout("unix", strings.TrimPrefix(c.host, "unix://"), time.Second*10)
			}

			c.transport.DialContext = dial
			c.baseURL = "http://d"
			return
		}

		c.baseURL = "http://" + strings.TrimPrefix(c.host, "tcp://")
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		command.Run(args)
	}

	c.rootCmd.AddCommand(cmd)
}

// NewTableDisplay create a Display instance, use to format output with table.
func (c *Cli) NewTableDisplay() *Display {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, c.padding, ' ', 0)
	return &Display{w}
}

// Print output the obj's fields.
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

// NewRequest create a HTTPRequest instance.
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

	r, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to new http request")
	}
	return &Request{req: r, cli: c.httpcli}, nil
}

// NewGetRequest create a http get request.
func (c *Cli) NewGetRequest(path string) (*Request, error) {
	return c.NewRequest(http.MethodGet, path, nil)
}

// NewPostRequest create a http post request.
func (c *Cli) NewPostRequest(path string, obj interface{}, enc ...func(interface{}) (io.Reader, error)) (*Request, error) {
	return c.NewRequest(http.MethodPost, path, obj, enc...)
}

// NewDeleteRequest create a http delete request.
func (c *Cli) NewDeleteRequest(path string, obj interface{}, enc ...func(interface{}) (io.Reader, error)) (*Request, error) {
	return c.NewRequest(http.MethodDelete, path, obj, enc...)
}
