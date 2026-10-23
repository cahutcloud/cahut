package main

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/knock"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/web"
	"cahut.cloud/internal/window"
)

// sample is what the reference describes, read from a new server with one
// Cahut.
type sample struct {
	instructions string
	tools        []*mcp.Tool
	cahutOps     []operation.Info
	serverOps    []operation.Info
}

// The Cahut of the sample and its person. The MCP endpoint of a Cahut
// refuses other names than localhost on a loopback address, where the
// sample is served.
const (
	sampleDomain = "localhost"
	sampleHandle = "reader"
)

// admin is the actor of the commands that set the sample up.
var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

// inspect creates a server in dir with one Cahut, a person and a token for
// their AI, then connects to the Cahut's MCP endpoint as that AI.
func inspect(ctx context.Context, dir string) (sample, error) {
	srv, err := server.Open(ctx, server.Config{DataDir: dir})
	if err != nil {
		return sample{}, err
	}
	defer srv.Close()
	if _, err := operation.Do(ctx, srv.Registry, srv.AddDomain, admin, server.AddDomainInput{Domain: sampleDomain}); err != nil {
		return sample{}, err
	}
	token, cahutOps, err := createPerson(ctx, srv)
	if err != nil {
		return sample{}, err
	}
	s := sample{cahutOps: cahutOps, serverOps: srv.Registry.Ops()}

	w := web.New(srv, web.Config{}, slog.New(slog.DiscardHandler))
	defer w.Close()
	hs := httptest.NewServer(w)
	defer hs.Close()
	// Requests name the domain of the Cahut, and go to the test server.
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, hs.Listener.Addr().String())
		},
	}
	defer transport.CloseIdleConnections()

	client := mcp.NewClient(&mcp.Implementation{Name: "cahut-reference", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             "http://" + sampleDomain + "/mcp",
		HTTPClient:           &http.Client{Transport: bearer{token: token, next: transport}},
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		return sample{}, fmt.Errorf("connect over MCP: %w", err)
	}
	defer session.Close()
	s.instructions = session.InitializeResult().Instructions
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			return sample{}, fmt.Errorf("list the tools: %w", err)
		}
		s.tools = append(s.tools, tool)
	}
	slices.SortFunc(s.tools, func(a, b *mcp.Tool) int { return cmp.Compare(a.Name, b.Name) })
	return s, nil
}

// createPerson creates the person of the sample, an administrator of the
// Cahut so that their AI sees every tool, and a token for that AI. It returns
// the token and the operations of the Cahut.
func createPerson(ctx context.Context, srv *server.Server) (string, []operation.Info, error) {
	c, err := srv.OpenCahut(ctx, sampleDomain)
	if err != nil {
		return "", nil, err
	}
	defer c.Close()
	user := account.CreateUserInput{Handle: sampleHandle, Name: "Reference", Admin: true}
	if _, err := operation.Do(ctx, c.Registry, account.CreateUser, admin, user); err != nil {
		return "", nil, err
	}
	t, err := operation.Do(ctx, c.Registry, account.CreateToken, admin, account.CreateTokenInput{User: sampleHandle, Name: "Reference"})
	if err != nil {
		return "", nil, err
	}
	return t.Token, c.Registry.Ops(), nil
}

// bearer adds the token of an AI to every request.
type bearer struct {
	token string
	next  http.RoundTripper
}

func (b bearer) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+b.token)
	return b.next.RoundTrip(r)
}

func toolsPage(s sample) (page, error) {
	p := page{
		file:        "mcp-tools.md",
		title:       "MCP tools",
		description: "The tools that a Cahut offers to AIs over MCP, with their arguments, results and JSON Schemas.",
		order:       1,
	}
	m := &p.body
	m.block("Each Cahut serves MCP at `https://<domain>/mcp`, over Streamable HTTP. " +
		"An AI connects with a personal token of that Cahut, which it sends in the header `Authorization: Bearer <token>`: " +
		"see [Connect an AI](/connect/claude-code/).")
	m.block("The tools are the operations of the Cahut that are open to AIs ([Operations](/reference/operations/)). " +
		"An AI sees the tools that the person it acts for may use: some are for the AIs of administrators of the Cahut only. " +
		"The operations that prove who someone is, such as signing in, are for people only. " +
		"Every call goes through the registry of operations, which checks it, runs it and writes it to the Logbook, " +
		"reads included. Arguments with unknown fields are refused.")
	m.block("A tool returns its result as structured content, and as the same JSON in a text block. " +
		"When a call fails because of its arguments, the error says what to change. Any other error reads " +
		"only as an internal error: the administrator of the Cahut finds the details in its Logbook.")

	m.block("## Instructions")
	m.block("The Cahut gives these instructions to the AI when it connects:")
	m.block("> " + text(s.instructions))

	m.block("## Limits")
	m.table([]string{"Limit", "Value"}, [][]string{
		{"Files in one Window", strconv.Itoa(window.MaxFiles)},
		{"Size of the files of one Window, decoded", mebibytes(window.MaxSize)},
		{"Size of one request to " + code("/mcp"), "about " + mebibytes(web.MaxRequestSize) + ", room for the files in base64"},
	})

	m.block("## Tools")
	var rows [][]string
	for _, t := range s.tools {
		rows = append(rows, []string{"[" + code(t.Name) + "](#" + t.Name + ")", text(t.Description)})
	}
	m.table([]string{"Tool", "What it does"}, rows)

	ops := map[string]operation.Info{}
	for _, op := range s.cahutOps {
		ops[op.Name] = op
	}
	for _, t := range s.tools {
		m.block("## " + code(t.Name))
		m.block(text(t.Description))
		op := ops[t.Name]
		if op.Level == operation.CahutAdmin {
			m.block("Offered to the AIs of administrators of the Cahut only.")
		}
		switch {
		case op.KnockFirst:
			m.block("Knock first: the call of an AI makes a Knock, which the person it acts for approves " +
				"([Operations](/reference/operations/#knock-first)).")
		case op.KnocksWhen != "":
			m.block("Knock first when " + op.KnocksWhen + ": that call of an AI makes a Knock, which the person it acts for approves " +
				"([Operations](/reference/operations/#knock-first)).")
		}
		m.block("Annotations: " + annotations(t.Annotations) + ".")
		if err := describeSchema(m, "Arguments", "Required", t.InputSchema, "This tool takes no arguments."); err != nil {
			return page{}, fmt.Errorf("arguments of %s: %w", t.Name, err)
		}
		if err := describeSchema(m, "Result", "Always present", t.OutputSchema, "This tool returns no structured result."); err != nil {
			return page{}, fmt.Errorf("result of %s: %w", t.Name, err)
		}
	}
	return p, nil
}

// annotations lists the hints that a tool gives to its clients.
func annotations(a *mcp.ToolAnnotations) string {
	if a == nil {
		return "none"
	}
	hints := []string{code("readOnlyHint: " + strconv.FormatBool(a.ReadOnlyHint))}
	if a.DestructiveHint != nil {
		hints = append(hints, code("destructiveHint: "+strconv.FormatBool(*a.DestructiveHint)))
	}
	if a.IdempotentHint {
		hints = append(hints, code("idempotentHint: true"))
	}
	if a.OpenWorldHint != nil {
		hints = append(hints, code("openWorldHint: "+strconv.FormatBool(*a.OpenWorldHint)))
	}
	return strings.Join(hints, ", ")
}

// describeSchema writes a section on a JSON Schema of a tool: a table of its
// fields, then the schema itself. present heads the column that says
// whether a field is required.
func describeSchema(m *markdown, title, present string, v any, empty string) error {
	m.block("### " + title)
	if v == nil {
		m.block(empty)
		return nil
	}
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	var s schema
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	if len(s.Properties) == 0 {
		m.block(empty)
	} else {
		m.table([]string{"Field", "Type", present, "Description"}, s.rows(""))
	}
	m.details("JSON Schema", "json", string(raw))
	return nil
}

// schema is the part of a JSON Schema that the reference describes.
type schema struct {
	Type        json.RawMessage    `json:"type"`
	Description string             `json:"description"`
	Properties  map[string]*schema `json:"properties"`
	Required    []string           `json:"required"`
	Items       *schema            `json:"items"`
}

// rows lists the fields of s, required ones first, with those of their
// objects after them. prefix starts their paths.
func (s *schema) rows(prefix string) [][]string {
	names := make([]string, 0, len(s.Properties))
	for name := range s.Properties {
		names = append(names, name)
	}
	required := func(name string) bool { return slices.Contains(s.Required, name) }
	slices.SortFunc(names, func(a, b string) int {
		if ra, rb := required(a), required(b); ra != rb {
			if ra {
				return -1
			}
			return 1
		}
		return cmp.Compare(a, b)
	})
	var rows [][]string
	for _, name := range names {
		f := s.Properties[name]
		path := prefix + name
		rows = append(rows, []string{code(path), f.typeName(), yesNo(required(name)), text(f.Description)})
		switch {
		case len(f.Properties) > 0:
			rows = append(rows, f.rows(path+".")...)
		case f.Items != nil && len(f.Items.Properties) > 0:
			rows = append(rows, f.Items.rows(path+"[].")...)
		}
	}
	return rows
}

// typeName names the type of s, such as "string or null" or "array of
// objects".
func (s *schema) typeName() string {
	var types []string
	var one string
	if json.Unmarshal(s.Type, &one) == nil {
		types = []string{one}
	} else if json.Unmarshal(s.Type, &types) != nil {
		return "any"
	}
	// null last, as in "string or null".
	slices.SortStableFunc(types, func(a, b string) int {
		return cmp.Compare(boolInt(a == "null"), boolInt(b == "null"))
	})
	for i, t := range types {
		if t == "array" && s.Items != nil {
			types[i] = "array of " + s.Items.typeName() + "s"
		}
	}
	return strings.Join(types, " or ")
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// mebibytes formats a size in MiB, with one decimal when it needs one.
func mebibytes(n int) string {
	if n%(1<<20) == 0 {
		return strconv.Itoa(n>>20) + " MiB"
	}
	return strconv.FormatFloat(float64(n)/(1<<20), 'f', 1, 64) + " MiB"
}

func operationsPage(s sample) page {
	p := page{
		file:        "operations.md",
		title:       "Operations",
		description: "The operations of a Cahut and of its server: the only way to act on them.",
		order:       3,
	}
	m := &p.body
	m.block("Every action on a Cahut goes through a registry of operations. The MCP tools and the command line are " +
		"thin layers over it: the registry checks who calls, runs the operation and writes the Logbook. " +
		"Each Cahut has its own registry and its own Logbook, and so does the server that holds the Cahuts.")
	tools := map[string]bool{}
	for _, t := range s.tools {
		tools[t.Name] = true
	}
	m.block("Each operation states who may call it: any person of the Cahut, its administrators, or the administrators " +
		"of the server, who use the command line. An AI acts with the rights of the person it acts for.")
	head := []string{"Operation", "What it does", "Who calls it", "Changes data", "Offered to AIs"}
	m.block("## Operations of a Cahut")
	m.table(head, opRows(s.cahutOps, tools))
	m.block("## Operations of the server")
	m.block("They are not offered over MCP.")
	m.table(head, opRows(s.serverOps, nil))

	m.block("## Knock first")
	var always, when []string
	for _, op := range slices.Concat(s.cahutOps, s.serverOps) {
		switch {
		case op.KnockFirst:
			always = append(always, code(op.Name))
		case op.KnocksWhen != "":
			when = append(when, code(op.Name)+", when "+op.KnocksWhen)
		}
	}
	m.block(fmt.Sprintf("An operation that cannot be undone, or that changes what a person may do, knocks first. "+
		"When an AI calls it, the call does not run: Cahut checks it, then makes a Knock, "+
		"which waits %d hours for the person the AI acts for to approve or deny it in the console. "+
		"Approving runs the call, as the AI. A person who calls the operation in the console acts at once, after a confirmation.",
		int(knock.Lifetime.Hours())))
	if len(always) > 0 {
		m.block("These operations knock first: " + strings.Join(always, ", ") + ".")
	}
	if len(when) > 0 {
		m.block("These knock first for some calls only: " + strings.Join(when, "; ") + ".")
	}
	return p
}

func opRows(ops []operation.Info, tools map[string]bool) [][]string {
	var rows [][]string
	for _, op := range ops {
		name := code(op.Name)
		who := capitalize(op.Level.String())
		if op.Visitors {
			who = "Anyone, to sign in"
		}
		offered := "no"
		switch {
		case tools[op.Name] && op.Level == operation.CahutAdmin:
			offered = "to the AIs of administrators"
		case tools[op.Name]:
			offered = "yes"
		case op.PeopleOnly:
			offered = "no, people only"
		}
		if tools[op.Name] {
			name = "[" + name + "](/reference/mcp-tools/#" + op.Name + ")"
		}
		rows = append(rows, []string{name, text(op.Description), who, yesNo(!op.ReadOnly), offered})
	}
	return rows
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
