package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/oauth"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/version"
)

const instructions = `This Cahut keeps what AIs make, on the user's own server. ` +
	`Publish a document, deck, dashboard or site with publish_window: give its HTML and its other files, ` +
	`and you get back the address where it opens; give that address to the user. ` +
	`Make index.html a complete page that works on its own. ` +
	`Link its files with relative paths, such as css/deck.css, never /css/deck.css: ` +
	`a Window is served under the address of its Room. ` +
	`Pages run in a sandbox where localStorage, sessionStorage, IndexedDB and cookies throw errors: ` +
	`keep state in memory or in the address. ` +
	`Every call is recorded in the Logbook of the Cahut.`

// mcpHandler returns the MCP endpoint (Streamable HTTP) of the Cahut c, which
// asks for a personal token of c. It offers an AI the operations of c that
// its person may call, as tools.
func (s *Server) mcpHandler(c *cahut.Cahut) http.Handler {
	// One server for each set of tools: those of a member or of an
	// administrator, who may change things or only read. The tools of a
	// session follow the role its person had when it began. Should the role
	// change, the registry still checks every call.
	var servers [2][2]*mcp.Server
	for _, admin := range []bool{false, true} {
		for _, readOnly := range []bool{false, true} {
			servers[b2i(admin)][b2i(readOnly)] = s.mcpServer(c, admin, readOnly)
		}
	}
	server := func(r *http.Request) *mcp.Server {
		if info := auth.TokenInfoFromContext(r.Context()); info != nil {
			if owner, ok := info.Extra["owner"].(account.Owner); ok {
				return servers[b2i(owner.Admin)][b2i(owner.ReadOnly)]
			}
		}
		return servers[0][1]
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
	// A request without a valid token learns where to get one: the
	// metadata of the resource names the Cahut's authorization server.
	// Personal tokens do not expire.
	return auth.RequireBearerToken(verify, &auth.RequireBearerTokenOptions{
		ResourceMetadataURL:    s.resourceMetadataURL(c.Domain),
		AllowMissingExpiration: true,
	})(handler)
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// mcpServer returns an MCP server that offers, as tools, the operations of c
// for the AIs of a person who administers c or not, when they may change
// things or only read.
func (s *Server) mcpServer(c *cahut.Cahut, admin, readOnly bool) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{Name: "cahut", Version: version.Get().Version},
		&mcp.ServerOptions{Instructions: instructions})
	for _, info := range c.Registry.Ops() {
		if !info.ForAIs(admin, readOnly) {
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
	out, err := jsonschema.ForType(info.Output, nil)
	if err != nil {
		return nil, fmt.Errorf("output schema of %s: %w", info.Name, err)
	}
	destructive, openWorld := info.KnockFirst, false
	return &mcp.Tool{
		Name:         info.Name,
		Description:  info.Description,
		InputSchema:  in,
		OutputSchema: out,
		Annotations: &mcp.ToolAnnotations{
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
