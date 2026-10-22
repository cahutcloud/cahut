// Command reference writes the reference pages of docs.cahut.cloud from the
// code: the MCP tools as the MCP endpoint of a Cahut serves them to an AI,
// the operations of the registries, and the commands as the command line
// describes them. It owns the folder it writes to and removes the other
// pages there.
//
//	go run ./internal/reference sites/docs.cahut.cloud/src/content/docs/reference
//
// make docs-reference runs it, and CI checks that its output is committed.
package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// marker opens every page, after its front matter.
const marker = "<!-- Generated from the code by make docs-reference: do not edit. -->"

// environment lists the variables from which the command line reads
// defaults, which the reference must not show.
var environment = []string{"CAHUT_DATA", "CAHUT_ACME_EMAIL", "CAHUT_ACME_CA"}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run ./internal/reference <folder>")
		os.Exit(2)
	}
	for _, name := range environment {
		os.Unsetenv(name)
	}
	if err := run(context.Background(), os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "reference: %v\n", err)
		os.Exit(1)
	}
}

// run writes the reference into dir.
func run(ctx context.Context, dir string) error {
	tmp, err := os.MkdirTemp("", "cahut-reference-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	s, err := inspect(ctx, tmp)
	if err != nil {
		return err
	}
	tagline, cmds, err := describeCommands()
	if err != nil {
		return err
	}
	tools, err := toolsPage(s)
	if err != nil {
		return err
	}
	return write(dir, []page{tools, cliPage(tagline, cmds), operationsPage(s), apiPage(s)})
}

// page is a page of the reference.
type page struct {
	file        string
	title       string
	description string
	order       int // in the sidebar
	body        markdown
}

func (p page) render() []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "---\ntitle: %s\ndescription: %s\nsidebar:\n  order: %d\n---\n\n",
		strconv.Quote(p.title), strconv.Quote(p.description), p.order)
	b.WriteString(marker + "\n\n")
	b.WriteString(strings.TrimRight(p.body.String(), "\n") + "\n")
	return []byte(b.String())
}

// write writes pages into dir, which must exist, and removes the other
// pages there. The writes go through an os.Root, so that no page can land
// outside dir.
func write(dir string, pages []page) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	keep := map[string]bool{}
	for _, p := range pages {
		keep[p.file] = true
		if err := root.WriteFile(p.file, p.render(), 0o644); err != nil {
			return err
		}
	}
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return err
	}
	for _, e := range entries {
		ext := path.Ext(e.Name())
		if !keep[e.Name()] && (ext == ".md" || ext == ".mdx") {
			if err := root.Remove(e.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}

// markdown builds the body of a page, block by block.
type markdown struct{ strings.Builder }

// block writes lines as one block, followed by a blank line.
func (m *markdown) block(lines ...string) {
	for _, l := range lines {
		m.WriteString(l)
		m.WriteByte('\n')
	}
	m.WriteByte('\n')
}

// table writes a table.
func (m *markdown) table(header []string, rows [][]string) {
	lines := []string{row(header), row(slices.Repeat([]string{"---"}, len(header)))}
	for _, r := range rows {
		lines = append(lines, row(r))
	}
	m.block(lines...)
}

// code writes a fenced code block.
func (m *markdown) code(lang, text string) {
	m.block("```"+lang, strings.TrimRight(text, "\n"), "```")
}

// details writes a code block that opens on a click.
func (m *markdown) details(summary, lang, text string) {
	m.block("<details>", "<summary>"+summary+"</summary>")
	m.code(lang, text)
	m.block("</details>")
}

func row(cells []string) string {
	return "| " + strings.Join(cells, " | ") + " |"
}

// flagName finds the flags that a text names, such as --dev-addr.
var flagName = regexp.MustCompile(`(^|[\s(])(--[a-z][a-z0-9-]*)`)

// text escapes s for running text and table cells, and sets the flags it
// names as code, which the docs would otherwise typeset with a dash.
func text(s string) string {
	s = strings.NewReplacer(`\`, `\\`, "|", `\|`, "<", "&lt;", ">", "&gt;", "`", "\\`", "*", `\*`).Replace(s)
	return flagName.ReplaceAllString(s, "$1`$2`")
}

// code formats s as inline code.
func code(s string) string {
	return "`" + strings.ReplaceAll(s, "|", `\|`) + "`"
}

// sentence ends s with a period.
func sentence(s string) string {
	if strings.HasSuffix(s, ".") {
		return s
	}
	return s + "."
}
