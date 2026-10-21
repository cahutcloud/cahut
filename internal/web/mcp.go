package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/cahut"
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

// mcpHandler returns the MCP endpoint (Streamable HTTP) of the Cahut c. It
// offers every operation of c that is not reserved to the administrator as a
// tool, and asks for a personal token of c.
func (s *Server) mcpHandler(c *cahut.Cahut) http.Handler {
	srv := mcp.NewServer(&mcp.Implementation{Name: "cahut", Version: version.Get().Version},
		&mcp.ServerOptions{Instructions: instructions})
	for _, info := range c.Registry.Ops() {
		if info.AdminOnly {
			continue
		}
		tool, err := toolFor(info)
		if err != nil {
			panic(err) // a type the schema library cannot describe: a programming error
		}
		srv.AddTool(tool, s.callTool(c, info.Name))
	}
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv },
		&mcp.StreamableHTTPOptions{
			Logger:              s.log,
			SessionTimeout:      time.Hour,
			MaxRequestBodyBytes: MaxRequestSize, // the SDK's default, 4 MiB, is too small for a deck with images
		})
	verify := func(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		return s.verifyToken(ctx, c, token)
	}
	return auth.RequireBearerToken(verify, &auth.RequireBearerTokenOptions{AllowMissingExpiration: true})(handler)
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
	destructive, openWorld := info.Irreversible, false
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
			Client:  clientName(owner.TokenName, req.Session),
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

// verifyToken checks a personal token against the people of c only: a
// token of another Cahut is invalid here.
func (s *Server) verifyToken(ctx context.Context, c *cahut.Cahut, token string) (*auth.TokenInfo, error) {
	owner, err := account.Verify(ctx, c.DB, token, time.Now())
	if errors.Is(err, account.ErrInvalidToken) {
		return nil, fmt.Errorf("%w: %w", auth.ErrInvalidToken, err)
	}
	if err != nil {
		s.log.Error("verify a token", "cahut", c.Domain, "error", err)
		return nil, errors.New("internal error")
	}
	return &auth.TokenInfo{
		UserID: strconv.FormatInt(owner.UserID, 10),
		Extra:  map[string]any{"owner": owner},
	}, nil
}
