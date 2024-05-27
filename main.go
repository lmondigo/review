package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const usage = `review - Narrows linter reports to what has actually changed.

Usage:
    review [OPTIONS...] report_file

The report file must be in json format as generated by PMD.

Examples:
  Read diff from stdin and output the rules violations in the terminal
    git diff feature..main | review pmd-report.json

  Read diff from a file
    review -d path/to/changes.diff pmd-report.json

Return values:
  0 - no violation found
  1 - one or more violation found
  2 - error during command execution

Options:
  -d, --diff file
                Diff file location. Use '-' to read from
                stdin (defaults to stdin)
  -h, --help    Display this message
`

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr))
}

func run(r io.Reader, wout, werr io.Writer) int {
	var (
		diffFlag string
		helpFlag bool
	)

	flags := flag.NewFlagSet("review", flag.ContinueOnError)
	flags.StringVar(&diffFlag, "d", "", "")
	flags.StringVar(&diffFlag, "diff", "", "")
	flags.BoolVar(&helpFlag, "h", false, "")
	flags.BoolVar(&helpFlag, "help", false, "")
	flags.Usage = func() {}

	err := flags.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(werr, "%s\n", usage)
		fmt.Fprintf(werr, "error: %s\n", err)
		return 2
	}

	if helpFlag {
		fmt.Fprintf(wout, "%s\n", usage)
		return 0
	}

	if len(flags.Args()) != 1 {
		fmt.Fprintf(werr, "%s\n", usage)
		fmt.Fprint(werr, "error: no report file provided\n")
		return 2
	}
	reportFile := flags.Arg(0)

	var content contentChecker
	switch diffFlag {
	case "", "-":
		content, err = readDiff(r)
	default:
		content, err = readDiffFile(diffFlag)
	}
	if err != nil {
		fmt.Fprintf(werr, "error: could not read diff: %s\n", err)
		return 2
	}

	reporter := newTextReporter(wout)
	linter, err := readPMDFile(reportFile)
	if err != nil {
		fmt.Fprintf(werr, "error: could not read report: %s\n", err)
		return 2
	}

	cmd := command{
		content:  content,
		reporter: reporter,
		linter:   linter,
	}

	violationCount, err := cmd.run()
	if err != nil {
		fmt.Fprintf(werr, "error: command failed: %s\n", err)
		return 2
	}

	if violationCount > 0 {
		fmt.Fprintf(werr, "%d violations found.\n", violationCount)
		return 1
	}

	return 0
}
