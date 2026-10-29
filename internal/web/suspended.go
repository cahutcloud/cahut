package web

import (
	"net/http"
	"strings"

	"cahut.cloud/internal/housestyle"
	"cahut.cloud/internal/settings"
)

// A suspended Cahut keeps everything, and serves nothing of what it holds:
// its Windows, the pages of its Rooms, their frames, its Keys, its
// invitations and its house style answer with a page that says it is
// suspended. Its console still loads, for the administrators of its server,
// whose calls alone the registry runs; its MCP endpoint and its OAuth
// endpoints still answer, so that AIs stay connected while the registry
// refuses their calls. See docs/adr/0048-suspending-a-cahut.md.

// openWhileSuspended are the paths of port 443 that a suspended Cahut
// still serves, beside the files of the brand and its icons.
var openWhileSuspended = map[string]bool{
	"/":             true, // to the console
	"/mcp":          true,
	"/-/healthz":    true,
	"/-/pages.css":  true,
	"/-/pages.js":   true,
	"/-/frame.js":   true,
	"/oauth/token":  true,
	"/oauth/revoke": true,
	// A client that registers then meets the sign-in of its person, which
	// says that the Cahut is suspended.
	"/oauth/register":                           true,
	"/.well-known/oauth-protected-resource":     true,
	"/.well-known/oauth-protected-resource/mcp": true,
	"/.well-known/oauth-authorization-server":   true,
	"/.well-known/openid-configuration":         true,
}

// servedWhileSuspended reports whether a suspended Cahut serves the path p.
func servedWhileSuspended(p string) bool {
	return openWhileSuspended[p] || rootIcons[p] != "" || strings.HasPrefix(p, BrandPath)
}

// serveSuspended answers the request with the page of a suspended Cahut
// when the Cahut of the request is suspended, and reports whether it did.
// The page carries the brand of Cahut whatever the house style: the host
// speaks, not the organization. It names nothing that the Cahut holds,
// and offers no sign-in.
func (s *Server) serveSuspended(w http.ResponseWriter, r *http.Request) bool {
	c := siteOf(r).cahut
	suspension, err := settings.SuspensionOf(r.Context(), c.DB.Read)
	if err != nil {
		s.internal(w, "read the suspension of a Cahut", err)
		return true
	}
	if suspension == nil {
		return false
	}
	p := s.page(r, "", "Suspended")
	p.Brand = housestyle.Brand{Name: c.Domain, NeutralLook: true}
	p.SignIn = ""
	s.render(w, r, http.StatusServiceUnavailable, "suspended", p)
	return true
}
