package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/spf13/cobra"
)

// eventsDescription is used to describe events command in detail and auto generate command doc.
var eventsDescription = "events cli tool is used to subscribe pouchd events." +
	"We support filter parameter to filter some events that we care about or not."

// EventsCommand use to implement 'events' command.
type EventsCommand struct {
	baseCommand
	since  string
	until  string
	filter []string
}

// Init initialize events command.
func (e *EventsCommand) Init(c *Cli) {
	e.cli = c
	e.cmd = &cobra.Command{
		Use:   "events [OPTIONS]",
		Short: "Get real time events from the daemon",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return e.runEvents()
		},
		Example: eventsExample(),
	}
	e.addFlags()
}

// addFlags adds flags for specific command.
func (e *EventsCommand) addFlags() {
	flagSet := e.cmd.Flags()

	flagSet.StringVarP(&e.since, "since", "s", "", "Show all events created since timestamp")
	flagSet.StringVarP(&e.until, "until", "u", "", "Stream events until this timestamp")
	flagSet.StringSliceVarP(&e.filter, "filter", "f", []string{}, "Filter output based on conditions provided")
}

// runEvents is the entry of events command.
func (e *EventsCommand) runEvents() error {
	ctx := context.Background()
	apiClient := e.cli.Client()

	eventFilterArgs := filters.NewArgs()

	// TODO: parse params
	for _, f := range e.filter {
		var err error
		eventFilterArgs, err = filters.ParseFlag(f, eventFilterArgs)
		if err != nil {
			return err
		}
	}

	responseBody, err := apiClient.Events(ctx, e.since, e.until, eventFilterArgs)
	if err != nil {
		return err
	}

	return streamEvents(responseBody, os.Stdout)
}

// streamEvents decodes prints the incoming events in the provided output.
func streamEvents(input io.Reader, output io.Writer) error {
	return DecodeEvents(input, func(event types.EventsMessage, err error) error {
		if err != nil {
			return err
		}
		printOutput(event, output)
		return nil
	})
}

type eventProcessor func(event types.EventsMessage, err error) error

// printOutput prints all types of event information.
// Each output includes the event type, actor id, name and action.
// Actor attributes are printed at the end if the actor has any.
func printOutput(event types.EventsMessage, output io.Writer) {
	// skip empty event message
	if event == (types.EventsMessage{}) {
		return
	}

	if event.TimeNano != 0 {
		fmt.Fprintf(output, "%s ", time.Unix(0, event.TimeNano).Format(utils.RFC3339NanoFixed))
	} else if event.Time != 0 {
		fmt.Fprintf(output, "%s ", time.Unix(event.Time, 0).Format(utils.RFC3339NanoFixed))
	}

	id := ""
	if event.Actor != nil {
		id = event.Actor.ID
	}
	fmt.Fprintf(output, "%s %s %s", event.Type, event.Action, id)

	if event.Actor != nil && len(event.Actor.Attributes) > 0 {
		var attrs []string
		var keys []string
		for k := range event.Actor.Attributes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := event.Actor.Attributes[k]
			attrs = append(attrs, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Fprintf(output, " (%s)", strings.Join(attrs, ", "))
	}
	fmt.Fprint(output, "\n")
}

// DecodeEvents decodes event from input stream
func DecodeEvents(input io.Reader, ep eventProcessor) error {
	dec := json.NewDecoder(input)
	for {
		var event types.EventsMessage
		err := dec.Decode(&event)
		if err != nil && err == io.EOF {
			break
		}

		if procErr := ep(event, err); procErr != nil {
			return procErr
		}
	}
	return nil
}

func eventsExample() string {
	return `$ pouch events -s "2018-08-10T10:52:05"
	2018-08-10T10:53:15.071664386-04:00 volume create 9fff54f207615ccc5a29477f5ae2234c6b804ed8aad2f0dfc0dccb0cc69d4d12 (driver=local)
2018-08-10T10:53:15.091131306-04:00 container create f2b58eb6bc616d7a22bdb89de50b3f04e2c23134accdec1a9b9a7490d609d34c (image=registry.hub.docker.com/library/centos:latest, name=test)
2018-08-10T10:53:15.537704818-04:00 container start f2b58eb6bc616d7a22bdb89de50b3f04e2c23134accdec1a9b9a7490d609d34c (image=registry.hub.docker.com/library/centos:latest, name=test)`
}
