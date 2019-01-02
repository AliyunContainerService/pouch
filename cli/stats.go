package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

const (
	containerHeader     = "CONTAINER ID"
	containerNameHeader = "NAME"
	cpuPercHeader       = "CPU %"
	netIOHeader         = "NET I/O"
	blockIOHeader       = "BLOCK I/O"
	memPercHeader       = "MEM %"
	memUseHeader        = "MEM USAGE / LIMIT"
	pidsHeader          = "PIDS"
)

// statsDescription is used to describe stats command in detail and auto generate command doc.
var statsDescription = "stats command is to display a live stream of container(s) resource usage statistics"

// StatsCommand use to implement 'stats' command
type StatsCommand struct {
	baseCommand

	noStream bool
	//TODO: add more flags support
}

// Init initialize stats command.
func (stats *StatsCommand) Init(c *Cli) {
	stats.cli = c
	stats.cmd = &cobra.Command{
		Use:   "stats [OPTIONS] CONTAINER [CONTAINER...]",
		Short: "Display a live stream of container(s) resource usage statistics",
		Long:  statsDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return stats.runStats(args)
		},
		Example: statsExample(),
	}
	stats.addFlags()
}

// addFlags adds flags for specific command.
func (stats *StatsCommand) addFlags() {
	flagSet := stats.cmd.Flags()
	flagSet.BoolVar(&stats.noStream, "no-stream", false, "Disable streaming stats and only pull the first result")
}

// runStats is the entry of stats command.
func (stats *StatsCommand) runStats(args []string) error {
	ctx := context.Background()
	apiClient := stats.cli.Client()
	containers := args

	cStats := []*StatsEntryWithLock{}
	waitFirst := &sync.WaitGroup{}
	for _, name := range containers {
		s := &StatsEntryWithLock{StatsEntry: StatsEntry{container: name}}
		cStats = append(cStats, s)
		waitFirst.Add(1)
		go collect(ctx, s, apiClient, !stats.noStream, waitFirst)
	}

	// before print to screen, make sure each container get at least one valid stat data
	waitFirst.Wait()

	// do a quick scan in case container not found and so on
	var errs []string
	for _, c := range cStats {
		err := c.GetError()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	cleanScreen := func() {
		if !stats.noStream {
			fmt.Fprint(os.Stdout, "\033[2J")
			fmt.Fprint(os.Stdout, "\033[H")
		}
	}

	display := stats.cli.NewTableDisplay()
	displayHead := []string{containerHeader, containerNameHeader, cpuPercHeader, memPercHeader,
		memUseHeader, netIOHeader, blockIOHeader, pidsHeader}

	for range time.Tick(500 * time.Millisecond) {
		cleanScreen()
		ccstats := []StatsEntry{}

		// get snapshot of each container stats
		for _, c := range cStats {
			err := c.GetError()
			if err != nil {
				return err
			}
			ccstats = append(ccstats, c.GetStatsEntry())
		}

		display.AddRow(displayHead)
		// display the stats of each container
		for _, c := range ccstats {
			displayLine := []string{c.ID(), c.Name(), c.CPUPerc(), c.MemPerc(),
				c.MemUsage(), c.NetIO(), c.BlockIO(), c.PIDs()}
			display.AddRow(displayLine)
		}

		display.Flush()

		if stats.noStream {
			break
		}
	}

	return nil
}

// statsExample shows examples in stats command, and is used in auto-generated cli docs.
func statsExample() string {
	return `$ pouch stats b25ae a0067
CONTAINER ID        NAME                       CPU %               MEM USAGE / LIMIT     MEM %               NET I/O             BLOCK I/O           PIDS
b25ae88e5b70        naughty_goldwasser         0.11%               2.559MiB / 15.23GiB   0.02%               7.32kB / 0B         0B / 0B             4
a00670c2bdff        xenodochial_varahamihira   0.11%               2.887MiB / 15.23GiB   0.02%               13.3kB / 0B         14.7MB / 0B         4
`
}
