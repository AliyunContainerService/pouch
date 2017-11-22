package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alibaba/pouch/ctrd"

	"github.com/containerd/containerd/progress"
	"github.com/spf13/cobra"
)

// PullCommand use to implement 'pull' command, it download image.
type PullCommand struct {
	baseCommand
}

// Init initialize pull command.
func (p *PullCommand) Init(c *Cli) {
	p.cli = c

	p.cmd = &cobra.Command{
		Use:   "pull [image]",
		Short: "Pull use to download image from repository",
		Args:  cobra.MinimumNArgs(1),
	}
}

// Run is the entry of pull command.
func (p *PullCommand) Run(args []string) {
	name, tag := parseNameTag(args[0])

	apiClient := p.cli.Client()
	responseBody, err := apiClient.ImagePull(name, tag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to pull image: %v \n", err)
		return
	}
	defer responseBody.Close()

	renderOutput(responseBody)
}

// renderOutput draws the commandline output via api response.
func renderOutput(responseBody io.ReadCloser) {
	var (
		start = time.Now()
		fw    = progress.NewWriter(os.Stdout)
	)

	dec := json.NewDecoder(responseBody)
	if _, err := dec.Token(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read the opening token: %v", err)
		return
	}

	for dec.More() {
		var objs []ctrd.ProgressInfo

		tw := tabwriter.NewWriter(fw, 1, 8, 1, ' ', 0)

		if err := dec.Decode(&objs); err != nil {
			fmt.Fprintf(os.Stderr, "failed to decode: %v \n", err)
			return
		}

		display(tw, objs, start)

		tw.Flush()
		fw.Flush()
	}

	if _, err := dec.Token(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read the closing token: %v", err)
		return
	}
}

func display(w io.Writer, statuses []ctrd.ProgressInfo, start time.Time) {
	var total int64
	for _, status := range statuses {
		total += status.Offset
		switch status.Status {
		case "downloading", "uploading":
			var bar progress.Bar
			if status.Total > 0.0 {
				bar = progress.Bar(float64(status.Offset) / float64(status.Total))
			}
			fmt.Fprintf(w, "%s:\t%s\t%40r\t%8.8s/%s\t\n",
				status.Ref,
				status.Status,
				bar,
				progress.Bytes(status.Offset), progress.Bytes(status.Total))

		case "resolving", "waiting":
			bar := progress.Bar(0.0)
			fmt.Fprintf(w, "%s:\t%s\t%40r\t\n",
				status.Ref,
				status.Status,
				bar)

		default:
			bar := progress.Bar(1.0)
			fmt.Fprintf(w, "%s:\t%s\t%40r\t\n",
				status.Ref,
				status.Status,
				bar)
		}
	}

	fmt.Fprintf(w, "elapsed: %-4.1fs\ttotal: %7.6v\t(%v)\t\n",
		time.Since(start).Seconds(),
		progress.Bytes(total),
		progress.NewBytesPerSecond(total, time.Since(start)))
}

// parseNameTag parses input arg and gets image name and image tag.
func parseNameTag(input string) (string, string) {
	fields := strings.SplitN(input, ":", 2)

	var name, tag string

	name = fields[0]

	if len(fields) == 1 {
		tag = "latest"
	} else if len(fields) == 2 {
		tag = fields[1]
	}

        if strings.Contains(name, "/") {
                count := strings.Count(name, "/")
                if count == 1 {
                  name = "docker.io/" + name
                }
        } else {
            name = "docker.io/library/" + name
        }
       
	return name, tag
}
