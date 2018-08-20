package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pouch/pkg/utils"

	"github.com/docker/docker/pkg/stringid"
	"github.com/spf13/cobra"
)

// historyDescription is used to describe history command in detail and auto generate command doc.
var historyDescription = "Return the history information about image"

// HistoryCommand is used to implement 'image history' command.
type HistoryCommand struct {
	baseCommand

	// flags for history command
	flagHuman   bool
	flagQuiet   bool
	flagNoTrunc bool
}

// Init initialize "image history" command.
func (h *HistoryCommand) Init(c *Cli) {
	h.cli = c
	h.cmd = &cobra.Command{
		Use:   "history [OPTIONS] IMAGE",
		Short: "Display history information on image",
		Long:  historyDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.runHistory(args)
		},
		Example: h.example(),
	}
	h.addFlags()
}

// addFlags adds flags for specific command.
func (h *HistoryCommand) addFlags() {
	flagSet := h.cmd.Flags()
	flagSet.BoolVar(&h.flagHuman, "human", true, "Print information in human readable format")
	flagSet.BoolVarP(&h.flagQuiet, "quiet", "q", false, "Only show image numeric ID")
	flagSet.BoolVar(&h.flagNoTrunc, "no-trunc", false, "Do not truncate output")
}

// runHistory is used to get history of an image.
func (h *HistoryCommand) runHistory(args []string) error {
	name := args[0]

	ctx := context.Background()
	apiClient := h.cli.Client()

	history, err := apiClient.ImageHistory(ctx, name)
	if err != nil {
		return err
	}

	display := h.cli.NewTableDisplay()
	if h.flagQuiet {
		for _, entry := range history {
			if h.flagNoTrunc {
				display.AddRow([]string{entry.ID})
			} else {
				display.AddRow([]string{stringid.TruncateID(entry.ID)})
			}
		}
		display.Flush()
		return nil
	}

	var (
		imageID   string
		createdBy string
		created   string
		size      string
	)

	display.AddRow([]string{"IMAGE", "CREATED", "CREATED BY", "SIZE", "COMMENT"})
	for _, entry := range history {
		imageID = entry.ID
		createdBy = strings.Replace(entry.CreatedBy, "\t", " ", -1)
		if !h.flagNoTrunc {
			createdBy = ellipsis(createdBy, 45)
			imageID = stringid.TruncateID(entry.ID)
		}

		if h.flagHuman {
			created, err = utils.FormatTimeInterval(entry.Created)
			if err != nil {
				return err
			}
			created = created + " ago"
			size = utils.FormatSize(entry.Size)
		} else {
			created = time.Unix(0, entry.Created).Format(time.RFC3339)
			size = strconv.FormatInt(entry.Size, 10)
		}

		display.AddRow([]string{imageID, created, createdBy, size, entry.Comment})
	}
	display.Flush()
	return nil
}

// example shows examples in history command, and is used in auto-generated cli docs.
func (h *HistoryCommand) example() string {
	return `pouch history busybox:latest
IMAGE          CREATED      CREATED BY                                      SIZE        COMMENT
e1ddd7948a1c   1 week ago   /bin/sh -c #(nop)  CMD ["sh"]                   0.00 B
<missing>      1 week ago   /bin/sh -c #(nop) ADD file:96fda64a6b725d4...   716.06 KB  `
}

// ellipsis truncates a string to fit within maxlen, and appends ellipsis (...).
// For maxlen of 3 and lower, no ellipsis is appended.
func ellipsis(s string, maxlen int) string {
	r := []rune(s)
	if len(r) <= maxlen {
		return s
	}
	if maxlen <= 3 {
		return string(r[:maxlen])
	}
	return string(r[:maxlen-3]) + "..."
}
