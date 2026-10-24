package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/oauth"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/version"
)

// instructions guide the AIs that connect. What an AI needs to publish
// comes first, within the 512 characters that some clients, such as Codex,
// are sure to keep (docs/adr/0029-chatgpt-second-ai-client.md).
const instructions = `Cahut keeps what you make on the user's own server, as Windows. ` +
	`To publish a document, deck, dashboard or site: call get_house_style with its kind, then publish_window ` +
	`with a complete index.html, its files linked by relative paths (css/deck.css, never /css/deck.css), ` +
	`and a declaration of its scripts and of every origin it loads from or calls. Give the user the address it returns. ` +
	`A call that cannot be undone makes a Knock: give the user its link to approve it.

` +
	`House style: when it is active, start from its template, link the style sheet of the organization, /-/house-style.css, ` +
	`use its classes and variables, and follow its guidance; when it is inactive, make the Window in the style the user asks for. ` +
	`Only the house style and the deck engine are under /-/: a Window is served under the address of its Room. ` +
	`Pages can load and call nothing that their declaration does not list, and send forms only where it says. ` +
	`They run in a sandbox where localStorage, sessionStorage, IndexedDB and cookies throw errors: ` +
	`keep state in memory or in the address. ` +
	`To change a Window, publish its new version with update_window at the same address; rollback_window goes back to the previous one. ` +
	`A Window opens to the members of its Room until set_window_access opens it to the whole Cahut or to anyone with its Key; ` +
	`making it public waits for a person. ` +
	`invite_to_window sends its address to people by e-mail, with a personal link for a Window with a Key. ` +
	`Knock first: a call that cannot be undone, or that changes what a person may do, does not run at once; ` +
	`once the person has answered its Knock, get_knock gives the outcome. ` +
	`Every call is recorded in the Logbook of the Cahut.`

// An AI opens MCP sessions with its token: a burst of MCPSessionBurst, then
// one every MCPSessionEvery, so that one token cannot keep thousands of
// sessions in memory for an hour.
const (
	MCPSessionBurst = 20
	MCPSessionEvery = 3 * time.Second
)

// Requests to /mcp larger than BigRequest, such as a publication with
// images, are served BigRequests at a time: each takes several times its
// size in memory while it is decoded. The others wait their turn, up to
// BigRequestWait.
const (
	BigRequest     = 4 << 20
	BigRequests    = 2
	BigRequestWait = time.Minute
)

// mcpHandler returns the MCP endpoint (Streamable HTTP) of the Cahut c, which
// asks for a personal token of c. It offers an AI the operations of c that
// its person may call, as tools.
func (s *Server) mcpHandler(c *cahut.Cahut) http.Handler {
	// One server for each set of tools: those of a member or of an
	// administrator, who may change things or only read, in a Cahut that is
	// individual or not. Each is made when a session first needs it. The
	// tools of a session follow the role its person had and the kind the
	// Cahut had when it began. Should either change, the registry still
	// checks every call.
	var mu sync.Mutex
	servers := map[[3]bool]*mcp.Server{}
	serverFor := func(individual, admin, readOnly bool) *mcp.Server {
		mu.Lock()
		defer mu.Unlock()
		k := [3]bool{individual, admin, readOnly}
		if servers[k] == nil {
			servers[k] = s.mcpServer(c, individual, admin, readOnly)
		}
		return servers[k]
	}
	server := func(r *http.Request) *mcp.Server {
		individual, err := settings.IsIndividual(r.Context(), c.DB.Read)
		if err != nil {
			individual = true // the narrower set of tools
		}
		if info := auth.TokenInfoFromContext(r.Context()); info != nil {
			if owner, ok := info.Extra["owner"].(account.Owner); ok {
				return serverFor(individual, owner.Admin, owner.ReadOnly)
			}
		}
		return serverFor(individual, false, true)
	}
	handler := mcp.NewStreamableHTTPHandler(server,
		&mcp.StreamableHTTPOptions{
			Logger:              s.log,
			SessionTimeout:      time.Hour,
			MaxRequestBodyBytes: MaxRequestSize, // the SDK's default, 4 MiB, is too small for a deck with images
			// On a loopback address, the SDK's guard against DNS rebinding
			// accepts the host localhost only, and would turn away a Cahut at
			// beta.localhost. ServeHTTP guards instead, on every address: it
			// answers only the domains of the Cahuts.
			DisableLocalhostProtection: true,
		})
	verify := func(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		return s.verifyToken(ctx, c, token)
	}
	limited := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.Header.Get("Mcp-Session-Id") == "" && !s.mcpSessions.allow(sessionKey(c, r)) {
			w.Header().Set("Retry-After", "10")
			http.Error(w, "Too many MCP sessions opened with this token: wait a few seconds.", http.StatusTooManyRequests)
			return
		}
		if r.ContentLength > BigRequest || r.ContentLength < 0 && r.Method == http.MethodPost {
			timer := time.NewTimer(BigRequestWait)
			defer timer.Stop()
			select {
			case s.bigRequests <- struct{}{}:
				defer func() { <-s.bigRequests }()
			case <-timer.C:
				w.Header().Set("Retry-After", "30")
				http.Error(w, "The server is busy with other large requests: try again in a moment.", http.StatusServiceUnavailable)
				return
			case <-r.Context().Done():
				return
			}
		}
		handler.ServeHTTP(w, r)
	})
	// A request without a valid token learns where to get one: the
	// metadata of the resource names the Cahut's authorization server.
	// Personal tokens do not expire.
	return auth.RequireBearerToken(verify, &auth.RequireBearerTokenOptions{
		ResourceMetadataURL:    s.resourceMetadataURL(c.Domain),
		AllowMissingExpiration: true,
	})(limited)
}

// sessionKey names the token of a request to the MCP endpoint of c, which
// the bearer guard verified: a personal token, or a connection over OAuth.
func sessionKey(c *cahut.Cahut, r *http.Request) string {
	if info := auth.TokenInfoFromContext(r.Context()); info != nil {
		if owner, ok := info.Extra["owner"].(account.Owner); ok {
			if owner.Connection != 0 {
				return c.Domain + " c" + strconv.FormatInt(owner.Connection, 10)
			}
			return c.Domain + " t" + strconv.FormatInt(owner.TokenID, 10)
		}
	}
	return c.Domain + " " + clientKey(r)
}

// mcpServer returns an MCP server that offers, as tools, the operations of c
// for the AIs of a person who administers c or not, when they may change
// things or only read, in a Cahut that is individual or not.
func (s *Server) mcpServer(c *cahut.Cahut, individual, admin, readOnly bool) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{Name: "cahut", Version: version.Get().Version},
		&mcp.ServerOptions{Instructions: instructions})
	for _, info := range c.Registry.Ops() {
		if !info.ForAIsIn(individual, admin, readOnly) {
			continue
		}
		tool, err := toolFor(info)
		if err != nil {
			panic(err) // a type the schema library cannot describe: a programming error
		}
		srv.AddTool(tool, s.callTool(c, info.Name))
	}
	return srv
}

func toolFor(info operation.Info) (*mcp.Tool, error) {
	in, err := jsonschema.ForType(info.Input, nil)
	if err != nil {
		return nil, fmt.Errorf("input schema of %s: %w", info.Name, err)
	}
	destructive, openWorld := info.KnockFirst, info.OpenWorld
	description := info.Description
	switch {
	case info.KnockFirst:
		description += " Knock first: a person approves the call before it runs."
	case info.KnocksWhen != "":
		description += " Knock first when " + info.KnocksWhen + ": a person approves the call before it runs."
	}
	// No output schema: the results stay structured, and the list of tools
	// stays light for the clients that load it whole.
	return &mcp.Tool{
		Name:        info.Name,
		Title:       info.Title,
		Description: description,
		InputSchema: in,
		Annotations: &mcp.ToolAnnotations{
			Title:           info.Title,
			ReadOnlyHint:    info.ReadOnly,
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
		},
	}, nil
}

// callTool runs an operation of c for the AI that holds the request's
// token.
func (s *Server) callTool(c *cahut.Cahut, name string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner, ok := req.Extra.TokenInfo.Extra["owner"].(account.Owner)
		if !ok {
			return nil, errors.New("no token owner in the request")
		}
		actor := operation.Actor{
			Kind:    operation.AI,
			Via:     "mcp",
			UserID:  owner.UserID,
			Handle:  owner.Handle,
			TokenID: owner.TokenID,
			Conn:    owner.Connection,
			Client:  clientName(owner.TokenName, req.Session),

			ReadOnly: owner.ReadOnly,
			Rooms:    owner.Rooms,
		}
		out, err := c.Registry.Invoke(ctx, name, actor, req.Params.Arguments)
		if err != nil {
			msg := err.Error()
			if !operation.Public(err) {
				s.log.Error("tool call failed", "cahut", c.Domain, "tool", name, "user", owner.Handle, "error", err)
				msg = "Internal error. The administrator of the Cahut can find it in the Logbook."
			}
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: msg}}}, nil
		}
		text, err := json.Marshal(out)
		if err != nil {
			return nil, err
		}
		return &mcp.CallToolResult{
			Content:           []mcp.Content{&mcp.TextContent{Text: string(text)}},
			StructuredContent: out,
		}, nil
	}
}

// clientName names the AI for the Logbook: the name of its token, then the
// software it says it is, such as "Claude Code (claude-code 2.1.0)".
func clientName(token string, session *mcp.ServerSession) string {
	if session == nil {
		return token
	}
	p := session.InitializeParams()
	if p == nil || p.ClientInfo == nil || p.ClientInfo.Name == "" {
		return token
	}
	software := p.ClientInfo.Name
	if p.ClientInfo.Version != "" {
		software += " " + p.ClientInfo.Version
	}
	return token + " (" + software + ")"
}

// verifyToken checks a personal token, or an OAuth access token, against
// the people of c only: a token of another Cahut is invalid here.
func (s *Server) verifyToken(ctx context.Context, c *cahut.Cahut, token string) (*auth.TokenInfo, error) {
	var owner account.Owner
	var expires time.Time
	var err error
	if strings.HasPrefix(token, oauth.AccessPrefix) {
		owner, expires, err = c.OAuth.AccessOwner(ctx, token)
	} else {
		owner, err = account.Verify(ctx, c.DB, token, time.Now())
	}
	if errors.Is(err, account.ErrInvalidToken) {
		return nil, fmt.Errorf("%w: %w", auth.ErrInvalidToken, err)
	}
	if err != nil {
		s.log.Error("verify a token", "cahut", c.Domain, "error", err)
		return nil, errors.New("internal error")
	}
	return &auth.TokenInfo{
		UserID:     strconv.FormatInt(owner.UserID, 10),
		Expiration: expires,
		Extra:      map[string]any{"owner": owner},
	}, nil
}
