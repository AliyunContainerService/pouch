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
	fields := strings.Split(args[0], ":")

	var (
		name string
		tag  string
	)

	if len(fields) == 1 || len(fields) == 2 {
		name = fields[0]
	}
	if len(fields) == 2 {
		tag = fields[1]
	}

	req, err := p.cli.NewPostRequest(fmt.Sprintf("/images/create?fromImage=%s&tag=%s", name, tag), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to new request: %v \n", err)
		return
	}
	response := req.Send()

	if err := response.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to pull: %v", err)
		return
	}
	defer response.Close()

	var (
		start = time.Now()
		fw    = progress.NewWriter(os.Stdout)
	)

	dec := json.NewDecoder(response.Body)
	if _, err = dec.Token(); err != nil {
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

		p.display(tw, objs, start)

		tw.Flush()
		fw.Flush()
	}

	if _, err = dec.Token(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read the closing token: %v", err)
		return
	}
}

func (p *PullCommand) display(w io.Writer, statuses []ctrd.ProgressInfo, start time.Time) {
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
		// TODO(stevvooe): These calculations are actually way off.
		// Need to account for previously downloaded data. These
		// will basically be right for a download the first time
		// but will be skewed if restarting, as it includes the
		// data into the start time before.
		progress.Bytes(total),
		progress.NewBytesPerSecond(total, time.Since(start)))
}
