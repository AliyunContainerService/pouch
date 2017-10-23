package main

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// Display use to output something on screen with table format.
type Display struct {
	w *tabwriter.Writer
}

// AddRow add a row of data.
func (d *Display) AddRow(row []string) {
	fmt.Fprintln(d.w, strings.Join(row, "\t"))
}

// Flush output all rows on screen.
func (d *Display) Flush() error {
	return d.w.Flush()
}
