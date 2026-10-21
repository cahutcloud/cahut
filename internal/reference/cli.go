package main

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cahut.cloud/internal/cli"
)

// command is a command of the command line, as its help describes it.
type command struct {
	path     string // such as "cahut token create"
	summary  string
	synopsis string
	flags    []flagHelp
	subs     []command // for a group of commands, such as cahut token
}

// flagHelp is a flag, as the help of its command prints it.
type flagHelp struct {
	name  string // such as "--data"
	value string // the kind of value it takes, such as "string"; empty for a switch
	def   string // its default, empty when none
	usage string
}

// describeCommands reads the help of the command line and of each command.
func describeCommands() (tagline string, cmds []command, err error) {
	stdout, stderr, code := runCLI("help")
	if code != 0 {
		return "", nil, fmt.Errorf("cahut help: exit code %d: %s", code, stderr)
	}
	tagline, _, _ = strings.Cut(stdout, "\n")
	for _, e := range listedCommands(stdout) {
		if e.name == "help" {
			continue
		}
		c, err := describe([]string{e.name}, e.summary)
		if err != nil {
			return "", nil, err
		}
		cmds = append(cmds, c)
	}
	if len(cmds) == 0 {
		return "", nil, errors.New("cahut help lists no command")
	}
	return tagline, cmds, nil
}

// describe reads the help of the command at path args.
func describe(args []string, summary string) (command, error) {
	c := command{path: "cahut " + strings.Join(args, " "), summary: summary}
	stdout, stderr, code := runCLI(append(slices.Clip(args), "-h")...)
	if code != 0 {
		return command{}, fmt.Errorf("%s -h: exit code %d: %s", c.path, code, stderr)
	}
	if subs := listedCommands(stdout); subs != nil {
		for _, e := range subs {
			sub, err := describe(append(slices.Clip(args), e.name), e.summary)
			if err != nil {
				return command{}, err
			}
			c.subs = append(c.subs, sub)
		}
		return c, nil
	}
	var err error
	c.synopsis, c.flags, err = parseFlags(c.path, stderr)
	return c, err
}

func runCLI(args ...string) (stdout, stderr string, code int) {
	var out, errOut bytes.Buffer
	code = cli.Run(args, &out, &errOut)
	return out.String(), errOut.String(), code
}

type listed struct{ name, summary string }

// listedCommands reads the commands that a help lists after "Commands:".
// It returns nil when the help lists none.
func listedCommands(help string) []listed {
	_, list, ok := strings.Cut(help, "\nCommands:\n")
	if !ok {
		return nil
	}
	var entries []listed
	for _, line := range strings.Split(list, "\n") {
		name, summary, ok := strings.Cut(strings.TrimPrefix(line, "\t"), " ")
		if !ok || !strings.HasPrefix(line, "\t") {
			continue
		}
		entries = append(entries, listed{name: name, summary: strings.TrimSpace(summary)})
	}
	return entries
}

// parseFlags reads the help that the flag package prints for the command
// at path: a usage line, then each flag on a line, with its usage on the
// lines that follow.
func parseFlags(path, help string) (synopsis string, flags []flagHelp, err error) {
	lines := strings.Split(strings.TrimRight(help, "\n"), "\n")
	switch first := lines[0]; {
	case strings.HasPrefix(first, "Usage: "):
		synopsis = strings.TrimPrefix(first, "Usage: ")
	case first == "Usage of "+path+":":
	default:
		return "", nil, fmt.Errorf("%s -h: unexpected help %q", path, first)
	}
	for _, line := range lines[1:] {
		switch {
		case strings.HasPrefix(line, "  -"):
			head, usage, _ := strings.Cut(strings.TrimPrefix(line, "  -"), "\t")
			name, value, _ := strings.Cut(head, " ")
			flags = append(flags, flagHelp{name: "--" + name, value: value, usage: usage})
		case strings.HasPrefix(line, "    \t") && len(flags) > 0:
			f := &flags[len(flags)-1]
			f.usage = strings.TrimSpace(f.usage + " " + strings.TrimPrefix(line, "    \t"))
		default:
			return "", nil, fmt.Errorf("%s -h: unexpected line %q", path, line)
		}
	}
	for i := range flags {
		f := &flags[i]
		if usage, def, ok := strings.Cut(f.usage, " (default "); ok && strings.HasSuffix(def, ")") {
			f.usage, f.def = usage, strings.TrimSuffix(def, ")")
			if unquoted, err := strconv.Unquote(f.def); err == nil {
				f.def = unquoted
			}
		}
	}
	if synopsis == "" {
		synopsis = path
		if len(flags) > 0 {
			synopsis += " [flags]"
		}
	}
	return synopsis, flags, nil
}

func cliPage(tagline string, cmds []command) page {
	p := page{
		file:        "cli.md",
		title:       "Command line",
		description: "Every command of the cahut binary, with its flags, as the binary describes them.",
		order:       2,
	}
	m := &p.body
	m.block(text(tagline) + " The `cahut` binary serves the Cahuts of a server, and administers them from its command line.")
	m.code("text", "cahut <command> [arguments]")
	m.block("Run `cahut help` for the list of commands, and add `-h` to a command for its flags. " +
		"A flag takes one dash or two: `-data` and `--data` are the same. " +
		"On a server, run the commands as the service user, with the full path of the binary: " +
		"see [Administration](/admin/commands/).")

	var rows [][]string
	var leaves func(cmds []command)
	leaves = func(cmds []command) {
		for _, c := range cmds {
			if c.subs != nil {
				leaves(c.subs)
				continue
			}
			rows = append(rows, []string{"[" + code(c.path) + "](#" + anchor(c.path) + ")", text(c.summary)})
		}
	}
	leaves(cmds)
	m.table([]string{"Command", "What it does"}, rows)

	for _, c := range cmds {
		writeCommand(m, c, "##")
	}
	return p
}

func writeCommand(m *markdown, c command, level string) {
	m.block(level + " " + code(c.path))
	m.block(text(sentence(c.summary)))
	for _, sub := range c.subs {
		writeCommand(m, sub, level+"#")
	}
	if c.subs != nil {
		return
	}
	m.code("text", c.synopsis)
	if len(c.flags) == 0 {
		m.block("This command takes no flags.")
		return
	}
	var rows [][]string
	for _, f := range c.flags {
		name := f.name
		if f.value != "" {
			name += " " + f.value
		}
		def := ""
		if f.def != "" {
			def = code(f.def)
		}
		rows = append(rows, []string{code(name), def, text(f.usage)})
	}
	m.table([]string{"Flag", "Default", "Description"}, rows)
}

// anchor returns the id that the docs give to a heading made of s.
func anchor(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "-")
}
