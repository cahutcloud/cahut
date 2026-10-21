// Package cli implements the cahut command line.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"cahut.cloud/internal/version"
)

// Exit codes returned by Run.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

type command struct {
	name    string
	summary string
	run     func(args []string, stdout, stderr io.Writer) int
}

var commands = []command{
	{name: "version", summary: "Print the version of this binary", run: runVersion},
}

// Run runs the command named by args[0] and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return exitUsage
	}
	name := args[0]
	switch name {
	case "help", "-h", "-help", "--help":
		usage(stdout)
		return exitOK
	case "-version", "--version":
		name = "version"
	}
	for _, c := range commands {
		if c.name == name {
			return c.run(args[1:], stdout, stderr)
		}
	}
	fmt.Fprintf(stderr, "cahut: unknown command %q\n\n", name)
	usage(stderr)
	return exitUsage
}

func usage(w io.Writer) {
	fmt.Fprint(w, "Cahut is the home for everything your AIs make.\n\n")
	fmt.Fprint(w, "Usage:\n\n\tcahut <command> [arguments]\n\nCommands:\n\n")
	for _, c := range commands {
		fmt.Fprintf(w, "\t%-10s %s\n", c.name, c.summary)
	}
	fmt.Fprintf(w, "\t%-10s %s\n", "help", "Print this help")
}

func runVersion(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("cahut version", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(stderr, "cahut version: unexpected arguments")
		return exitUsage
	}
	if _, err := fmt.Fprintln(stdout, version.Get()); err != nil {
		return exitError
	}
	return exitOK
}
