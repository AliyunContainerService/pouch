package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/credential"
	"github.com/alibaba/pouch/pkg/jsonstream"
	"github.com/alibaba/pouch/pkg/reference"

	"github.com/containerd/containerd/progress"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// pullDescription is used to describe pull command in detail and auto generate command doc.
var pullDescription = "Pull an image or a repository from a registry. " +
	"Most of your images will be created on top of a base image from the registry. " +
	"So, you can pull and try prebuilt images contained by registry without needing to define and configure your own."

// PullCommand use to implement 'pull' command, it download image.
type PullCommand struct {
	baseCommand
}

// Init initialize pull command.
func (p *PullCommand) Init(c *Cli) {
	p.cli = c

	p.cmd = &cobra.Command{
		Use:   "pull IMAGE",
		Short: "Pull an image from registry",
		Long:  pullDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runPull(args)
		},
		Example: pullExample(),
	}
	p.addFlags()
}

// addFlags adds flags for specific command.
func (p *PullCommand) addFlags() {
	// TODO: add flags here
}

// runPull is the entry of pull command.
func (p *PullCommand) runPull(args []string) error {
	return pullMissingImage(context.Background(), p.cli.Client(), args[0], true)
}

func fetchRegistryAuth(serverAddress string) string {
	authConfig, err := credential.Get(serverAddress)
	if err != nil || authConfig == (types.AuthConfig{}) {
		return ""
	}

	data, err := json.Marshal(authConfig)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(data)
}

// bufwriter defines interface which has Write and Flush behaviors.
type bufwriter interface {
	Write([]byte) (int, error)
	Flush() error
}

// showProgress shows pull progress status.
func showProgress(body io.ReadCloser) error {
	var (
		output bufwriter = bufio.NewWriter(os.Stdout)

		start      = time.Now()
		isTerminal = terminal.IsTerminal(int(os.Stdout.Fd()))
	)

	if isTerminal {
		output = progress.NewWriter(os.Stdout)
	}

	pos := make(map[string]int)
	status := []jsonstream.JSONMessage{}

	dec := json.NewDecoder(body)
	for {
		var (
			msg  jsonstream.JSONMessage
			msgs []jsonstream.JSONMessage
		)

		if err := dec.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		change := true
		if _, ok := pos[msg.ID]; !ok {
			status = append(status, msg)
			pos[msg.ID] = len(status) - 1
		} else {
			change = (status[pos[msg.ID]].Status != msg.Status)
			status[pos[msg.ID]] = msg
		}

		// only display the new status if the stdout is not terminal
		if !isTerminal {
			// if the status doesn't change, skip to avoid duplicate status
			if !change {
				continue
			}
			msgs = []jsonstream.JSONMessage{msg}
		} else {
			msgs = status
		}

		if err := displayImageReferenceProgress(output, isTerminal, msgs, start); err != nil {
			return fmt.Errorf("failed to display progress: %v", err)
		}

		if err := output.Flush(); err != nil {
			return fmt.Errorf("failed to display progress: %v", err)
		}
	}
	return nil
}

// displayImageReferenceProgress uses tabwriter to show current progress status.
func displayImageReferenceProgress(output io.Writer, isTerminal bool, msgs []jsonstream.JSONMessage, start time.Time) error {
	var (
		tw      = tabwriter.NewWriter(output, 1, 8, 1, ' ', 0)
		current = int64(0)
	)

	for _, msg := range msgs {
		if msg.Error != nil {
			return fmt.Errorf(msg.Error.Message)
		}

		if msg.Detail != nil {
			current += msg.Detail.Current
		}

		status := jsonstream.ProcessStatus(!isTerminal, msg)
		if _, err := fmt.Fprint(tw, status); err != nil {
			return err
		}
	}

	// no need to show the total information if the stdout is not terminal
	if isTerminal {
		_, err := fmt.Fprintf(tw, "elapsed: %-4.1fs\ttotal: %7.6v\t(%v)\t\n",
			time.Since(start).Seconds(),
			progress.Bytes(current),
			progress.NewBytesPerSecond(current, time.Since(start)))
		if err != nil {
			return err
		}
	}
	return tw.Flush()
}

// pullExample shows examples in pull command, and is used in auto-generated cli docs.
func pullExample() string {
	return `$ pouch images
IMAGE ID            IMAGE NAME                           SIZE
bbc3a0323522        docker.io/library/busybox:latest     703.14 KB
$ pouch pull docker.io/library/redis:alpine
$ pouch images
IMAGE ID            IMAGE NAME                           SIZE
bbc3a0323522        docker.io/library/busybox:latest     703.14 KB
0153c5db97e5        docker.io/library/redis:alpine       9.63 MB`
}

// pullMissingImage pull the image if it doesn't exist.
// When `force` is true, always pull the latest image instead of
// using the local version
func pullMissingImage(ctx context.Context, apiClient client.CommonAPIClient, image string, force bool) error {
	if !force {
		_, inspectError := apiClient.ImageInspect(ctx, image)
		if inspectError == nil {
			return nil
		}
		if err, ok := inspectError.(client.RespError); !ok {
			return inspectError
		} else if err.Code() != http.StatusNotFound {
			return inspectError
		}
	}

	namedRef, err := reference.Parse(image)
	if err != nil {
		return err
	}

	namedRef = reference.TrimTagForDigest(reference.WithDefaultTagIfMissing(namedRef))

	var name, tag string
	if reference.IsNameTagged(namedRef) {
		name, tag = namedRef.Name(), namedRef.(reference.Tagged).Tag()
	} else {
		name = namedRef.String()
	}

	responseBody, err := apiClient.ImagePull(ctx, name, tag, fetchRegistryAuth(namedRef.Name()))
	if err != nil {
		return fmt.Errorf("failed to pull image: %v", err)
	}
	defer responseBody.Close()

	return showProgress(responseBody)
}
