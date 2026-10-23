// Package cli implements the cahut command line.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

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
	{name: "serve", summary: "Serve the Cahuts of this server: MCP for AIs, and the Windows", run: runServe},
	{name: "domain", summary: "Manage the Cahuts of this server, one per domain", run: group("domain", domainCommands)},
	{name: "admin", summary: "Manage the administrators of the server", run: group("admin", adminCommands)},
	{name: "user", summary: "Manage the people of a Cahut", run: group("user", userCommands)},
	{name: "token", summary: "Manage the personal tokens of AIs", run: group("token", tokenCommands)},
	{name: "logbook", summary: "Print the latest entries of the Logbook", run: runLogbook},
	{name: "update", summary: "Install a new release, as root, and go back if it fails", run: runUpdate},
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
	list(w, commands)
	fmt.Fprintf(w, "\t%-10s %s\n", "help", "Print this help")
}

func list(w io.Writer, cmds []command) {
	for _, c := range cmds {
		fmt.Fprintf(w, "\t%-10s %s\n", c.name, c.summary)
	}
}

// group returns a command that runs one of subs, as in "cahut token create".
func group(name string, subs []command) func(args []string, stdout, stderr io.Writer) int {
	return func(args []string, stdout, stderr io.Writer) int {
		if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
			w := stderr
			if len(args) > 0 {
				w = stdout
			}
			fmt.Fprintf(w, "Usage:\n\n\tcahut %s <command> [flags]\n\nCommands:\n\n", name)
			list(w, subs)
			if len(args) == 0 {
				return exitUsage
			}
			return exitOK
		}
		for _, c := range subs {
			if c.name == args[0] {
				return c.run(args[1:], stdout, stderr)
			}
		}
		fmt.Fprintf(stderr, "cahut %s: unknown command %q\n", name, args[0])
		return exitUsage
	}
}

// newFlags returns the flag set of a command, which reports errors on
// stderr.
func newFlags(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("cahut "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

// parse parses args into fs, which takes flags only. When ok is false, the
// command stops with code.
func parse(fs *flag.FlagSet, args []string) (code int, ok bool) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK, false
		}
		return exitUsage, false
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(fs.Output(), "%s: unexpected argument %q\n", fs.Name(), fs.Arg(0))
		return exitUsage, false
	}
	return 0, true
}

// parseWithArgs parses args into fs, which takes flags and exactly n
// arguments, in any order. It returns the arguments. When ok is false, the
// command stops with code.
func parseWithArgs(fs *flag.FlagSet, args []string, n int) (rest []string, code int, ok bool) {
	for {
		if err := fs.Parse(args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil, exitOK, false
			}
			return nil, exitUsage, false
		}
		if fs.NArg() == 0 {
			break
		}
		rest = append(rest, fs.Arg(0))
		args = fs.Args()[1:]
	}
	if len(rest) != n {
		fs.Usage()
		return nil, exitUsage, false
	}
	return rest, 0, true
}

// dataFlag adds the --data flag, for the directory of the server's data.
func dataFlag(fs *flag.FlagSet) *string {
	dir := os.Getenv("CAHUT_DATA")
	if dir == "" {
		dir = "/var/lib/cahut"
	}
	return fs.String("data", dir, "directory of the server's data, which holds its Cahuts (env CAHUT_DATA)")
}

func runVersion(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("version", stderr)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	if _, err := fmt.Fprintln(stdout, version.Get()); err != nil {
		return exitError
	}
	return exitOK
}
