package main

import (
	"io"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

type textReporter struct {
	w io.Writer
}

func newTextReporter(w io.Writer) *textReporter {
	return &textReporter{
		w,
	}
}

func (tr textReporter) report(violations []violation) error {
	if len(violations) == 0 {
		// nothing to print
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"File", "Line", "Severity", "Rule", "Message"})
	for _, v := range violations {
		t.AppendRow([]interface{}{
			v.filename,
			v.line,
			v.severity,
			v.rule,
			v.message})
	}
	t.Render()

	return nil
}
