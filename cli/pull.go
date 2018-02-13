package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/alibaba/pouch/ctrd"
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
	namedRef, err := reference.ParseNamedReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to pull image: %v", err)
	}
	taggedRef := reference.WithDefaultTagIfMissing(namedRef).(reference.Tagged)

	ctx := context.Background()
	apiClient := p.cli.Client()
	responseBody, err := apiClient.ImagePull(ctx, taggedRef.Name(), taggedRef.Tag())
	if err != nil {
		return fmt.Errorf("failed to pull image: %v", err)
	}
	defer responseBody.Close()

	return showProgress(responseBody)
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

	dec := json.NewDecoder(body)
	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("failed to read the opening token: %v", err)
	}

	refStatus := make(map[string]string)
	for dec.More() {
		var infos []ctrd.ProgressInfo

		if err := dec.Decode(&infos); err != nil {
			return fmt.Errorf("failed to decode: %v", err)
		}

		// only display the new status if the stdout is not terminal
		if !isTerminal {
			newInfos := make([]ctrd.ProgressInfo, 0)
			for i, info := range infos {
				old, ok := refStatus[info.Ref]
				if !ok || info.Status != old {
					refStatus[info.Ref] = info.Status
					newInfos = append(newInfos, infos[i])
				}
			}

			infos = newInfos
		}

		if err := displayProgressInfos(output, isTerminal, infos, start); err != nil {
			return fmt.Errorf("failed to display progress: %v", err)
		}

		if err := output.Flush(); err != nil {
			return fmt.Errorf("failed to display progress: %v", err)
		}
	}

	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("failed to read the closing token: %v", err)
	}
	return nil
}

// displayProgressInfos uses tabwriter to show current progress info.
func displayProgressInfos(output io.Writer, isTerminal bool, infos []ctrd.ProgressInfo, start time.Time) error {
	var (
		tw    = tabwriter.NewWriter(output, 1, 8, 1, ' ', 0)
		total = int64(0)
	)

	for _, info := range infos {
		if info.ErrorMessage != "" {
			return fmt.Errorf(info.ErrorMessage)
		}

		total += info.Offset
		if _, err := fmt.Fprint(tw, formatProgressInfo(info, isTerminal)); err != nil {
			return err
		}
	}

	// no need to show the total information if the stdout is not terminal
	if isTerminal {
		_, err := fmt.Fprintf(tw, "elapsed: %-4.1fs\ttotal: %7.6v\t(%v)\t\n",
			time.Since(start).Seconds(),
			progress.Bytes(total),
			progress.NewBytesPerSecond(total, time.Since(start)))
		if err != nil {
			return err
		}
	}
	return tw.Flush()
}

// formatProgressInfo formats ProgressInfo into string.
func formatProgressInfo(info ctrd.ProgressInfo, isTerminal bool) string {
	if !isTerminal {
		return fmt.Sprintf("%s:\t%s\n", info.Ref, info.Status)
	}

	switch info.Status {
	case "downloading", "uploading":
		var bar progress.Bar
		if info.Total > 0.0 {
			bar = progress.Bar(float64(info.Offset) / float64(info.Total))
		}
		return fmt.Sprintf("%s:\t%s\t%40r\t%8.8s/%s\t\n",
			info.Ref,
			info.Status,
			bar,
			progress.Bytes(info.Offset), progress.Bytes(info.Total))

	case "resolving", "waiting":
		return fmt.Sprintf("%s:\t%s\t%40r\t\n",
			info.Ref,
			info.Status,
			progress.Bar(0.0))

	default:
		return fmt.Sprintf("%s:\t%s\t%40r\t\n",
			info.Ref,
			info.Status,
			progress.Bar(1.0))
	}
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
