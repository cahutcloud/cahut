// Command apitypes writes the TypeScript types of the console's API from the
// Go types of the registries: the input and output of every operation of a
// Cahut and of the server. The console calls only what they declare, so a
// renamed operation or field fails its build. See
// docs/adr/0014-console-http-api.md. It writes the codes of the refusals
// too, with the values of each, which the console says in the language of
// its person: see docs/adr/0037-refusals-with-a-code-for-the-console.md.
// And it writes the values that the console shares with the server, such as
// the share of a limit from which both warn that a Cahut is close to it.
//
//	go run ./internal/apitypes internal/console/ui/src/api.gen.ts
//
// make console-types runs it, and CI checks that its output is committed.
package main

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path"
	"reflect"
	"slices"
	"strings"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/settings"
	_ "cahut.cloud/internal/web" // it declares refusals of the API itself
)

// header opens the file.
const header = "// Generated from the Go types of the registries by make console-types: do not edit.\n" +
	"// See docs/adr/0014-console-http-api.md.\n"

// admin is the actor that sets up the server whose registries are read.
var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run ./internal/apitypes <file.ts>")
		os.Exit(2)
	}
	if err := run(context.Background(), os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "apitypes: %v\n", err)
		os.Exit(1)
	}
}

// run writes the types into file.
func run(ctx context.Context, file string) error {
	cahutOps, serverOps, err := registries(ctx)
	if err != nil {
		return err
	}
	return os.WriteFile(file, []byte(generate(cahutOps, serverOps, operation.Refusals())), 0o644) //nolint:gosec // a source file, readable by all
}

// registries returns the operations of a Cahut and of the server, read from
// a server created for the purpose.
func registries(ctx context.Context) (cahutOps, serverOps []operation.Info, err error) {
	dir, err := os.MkdirTemp("", "cahut-apitypes-")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(dir)
	srv, err := server.Open(ctx, server.Config{DataDir: dir})
	if err != nil {
		return nil, nil, err
	}
	defer srv.Close()
	if _, err := operation.Do(ctx, srv.Registry, srv.AddDomain, admin, server.AddDomainInput{Domain: "localhost"}); err != nil {
		return nil, nil, err
	}
	c, err := srv.OpenCahut(ctx, "localhost")
	if err != nil {
		return nil, nil, err
	}
	defer c.Close()
	return c.Registry.Ops(), srv.Registry.Ops(), nil
}

// generate returns the TypeScript file: an interface for each Go struct
// that the operations take or return, or that fills a refusal, then the
// operations of a Cahut and of the server, each with its input and output,
// the refusals, each with its values, and the values shared with the server.
func generate(cahutOps, serverOps []operation.Info, refusals []operation.RefusalInfo) string {
	g := &generator{names: map[reflect.Type]string{}}
	for _, op := range slices.Concat(cahutOps, serverOps) {
		g.collect(op.Input)
		g.collect(op.Output)
	}
	for _, r := range refusals {
		if r.Values != nil {
			g.collect(r.Values)
		}
	}
	g.name()

	var b strings.Builder
	b.WriteString(header)
	types := make([]reflect.Type, 0, len(g.names))
	for t := range g.names {
		types = append(types, t)
	}
	slices.SortFunc(types, func(a, b reflect.Type) int { return cmp.Compare(g.names[a], g.names[b]) })
	for _, t := range types {
		b.WriteString("\n")
		g.declare(&b, t)
	}
	g.operations(&b, "CahutOperations", "The operations of a Cahut, at POST /console/api/<operation> of its domain.", cahutOps)
	g.operations(&b, "ServerOperations", "The operations of the server, at POST /console/api/server/<operation>, for its administrators.", serverOps)
	g.refusals(&b, refusals)
	constants(&b)
	return b.String()
}

// constants writes the values that the console shares with the server.
func constants(b *strings.Builder) {
	fmt.Fprintf(b, "\n/** The share of a limit, in percent, from which the console warns that it is close, as the e-mails of the Cahut do. */\n"+
		"export const nearLimit = %d;\n", settings.NearPercent)
}

type generator struct {
	names map[reflect.Type]string // the TypeScript name of each struct type
}

var timeType = reflect.TypeFor[time.Time]()

// collect records the named struct types that t is or holds.
func (g *generator) collect(t reflect.Type) {
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
		g.collect(t.Elem())
	case reflect.Struct:
		if t == timeType {
			return
		}
		if t.Name() != "" {
			if _, ok := g.names[t]; ok {
				return
			}
			g.names[t] = t.Name()
		}
		for _, f := range fields(t) {
			g.collect(f.Type)
		}
	}
}

// name gives each struct type its Go name, prefixed with its package when
// two types share a name, as in RoomList and WindowList.
func (g *generator) name() {
	count := map[string]int{}
	for t := range g.names {
		count[t.Name()]++
	}
	for t := range g.names {
		if count[t.Name()] == 1 {
			continue
		}
		pkg := path.Base(t.PkgPath())
		pkg = strings.ToUpper(pkg[:1]) + pkg[1:]
		if !strings.HasPrefix(t.Name(), pkg) {
			g.names[t] = pkg + t.Name()
		}
	}
}

// field is a field of a struct, as JSON sees it.
type field struct {
	Name     string
	Type     reflect.Type
	Optional bool
	Doc      string
}

// fields returns the fields of t that JSON encodes, those of embedded
// structs included, in order.
func fields(t reflect.Type) []field {
	var out []field
	for f := range t.Fields() {
		if !f.IsExported() {
			continue
		}
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, opts, _ := strings.Cut(tag, ",")
		if f.Anonymous && name == "" {
			ft := f.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			out = append(out, fields(ft)...)
			continue
		}
		if name == "" {
			name = f.Name
		}
		out = append(out, field{Name: name, Type: f.Type, Optional: strings.Contains(opts, "omitempty"), Doc: f.Tag.Get("jsonschema")})
	}
	return out
}

// declare writes the declaration of the struct type t.
func (g *generator) declare(b *strings.Builder, t reflect.Type) {
	fs := fields(t)
	if len(fs) == 0 {
		fmt.Fprintf(b, "export type %s = Record<string, never>;\n", g.names[t])
		return
	}
	fmt.Fprintf(b, "export interface %s {\n", g.names[t])
	for _, f := range fs {
		if f.Doc != "" {
			fmt.Fprintf(b, "  /** %s */\n", strings.ReplaceAll(f.Doc, "*/", "* /"))
		}
		opt := ""
		if f.Optional {
			opt = "?"
		}
		fmt.Fprintf(b, "  %s%s: %s;\n", f.Name, opt, g.typeOf(f.Type))
	}
	b.WriteString("}\n")
}

// typeOf returns the TypeScript type of the Go type t.
func (g *generator) typeOf(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Pointer:
		return g.typeOf(t.Elem())
	case reflect.Slice, reflect.Array:
		return g.typeOf(t.Elem()) + "[]"
	case reflect.Map:
		return "Record<string, " + g.typeOf(t.Elem()) + ">"
	case reflect.Struct:
		if t == timeType {
			return "string" // RFC 3339, in UTC
		}
		if name, ok := g.names[t]; ok {
			return name
		}
		if len(fields(t)) == 0 {
			return "Record<string, never>"
		}
		var b strings.Builder
		b.WriteString("{ ")
		for _, f := range fields(t) {
			opt := ""
			if f.Optional {
				opt = "?"
			}
			fmt.Fprintf(&b, "%s%s: %s; ", f.Name, opt, g.typeOf(f.Type))
		}
		b.WriteString("}")
		return b.String()
	}
	return "unknown"
}

// refusals writes the interface Refusals, which maps the code of each
// refusal to the type of its values, or null when its message holds none.
func (g *generator) refusals(b *strings.Builder, refusals []operation.RefusalInfo) {
	b.WriteString("\n/** The refusals that the console says in the language of its person, by code, each with its values. */\n")
	b.WriteString("export interface Refusals {\n")
	for _, r := range refusals {
		values := "null"
		if r.Values != nil {
			values = g.typeOf(r.Values)
		}
		fmt.Fprintf(b, "  /** %s */\n", strings.ReplaceAll(r.Message, "*/", "* /"))
		fmt.Fprintf(b, "  %s: %s;\n", r.Code, values)
	}
	b.WriteString("}\n")
}

// operations writes the interface called name, which maps each operation
// of ops to its input and output.
func (g *generator) operations(b *strings.Builder, name, doc string, ops []operation.Info) {
	fmt.Fprintf(b, "\n/** %s */\nexport interface %s {\n", doc, name)
	for _, op := range ops {
		fmt.Fprintf(b, "  /** %s */\n", strings.ReplaceAll(op.Description, "*/", "* /"))
		fmt.Fprintf(b, "  %s: { input: %s; output: %s };\n", op.Name, g.typeOf(op.Input), g.typeOf(op.Output))
	}
	b.WriteString("}\n")
}
