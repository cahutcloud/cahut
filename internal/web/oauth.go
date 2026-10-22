package web

import (
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"

	"cahut.cloud/internal/oauth"
)

// Each Cahut is the OAuth authorization server of its MCP endpoint, on the
// console's port. See docs/adr/0016-oauth-for-ai-clients.md.

// Registrations of OAuth clients take no session: they are limited by
// address, to a burst of RegisterBurst, then one every RegisterEvery.
const (
	RegisterBurst = 10
	RegisterEvery = time.Minute
)

// resourceMetadataURL is the address of the protected resource metadata of
// the MCP endpoint of the Cahut at domain (RFC 9728).
func (s *Server) resourceMetadataURL(domain string) string {
	return s.cfg.BaseURL(domain) + ".well-known/oauth-protected-resource/mcp"
}

// serveResourceMetadata answers the metadata of the MCP endpoint, which
// names its authorization server.
func (s *Server) serveResourceMetadata(w http.ResponseWriter, r *http.Request) {
	o := siteOf(r).cahut.OAuth
	auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
		Resource:               o.Resource(),
		AuthorizationServers:   []string{o.Issuer()},
		ScopesSupported:        []string{oauth.Read, oauth.Write},
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "Cahut at " + siteOf(r).cahut.Domain,
		ResourceDocumentation:  "https://docs.cahut.cloud/connect/ai-clients/",
	}).ServeHTTP(w, r)
}

func (s *Server) serveAuthorizationMetadata(w http.ResponseWriter, r *http.Request) {
	siteOf(r).cahut.OAuth.ServeMetadata(w, r)
}

func (s *Server) serveToken(w http.ResponseWriter, r *http.Request) {
	siteOf(r).cahut.OAuth.ServeToken(w, r)
}

func (s *Server) serveRevoke(w http.ResponseWriter, r *http.Request) {
	siteOf(r).cahut.OAuth.ServeRevoke(w, r)
}

// serveRegister registers an OAuth client. Registration is open, as MCP
// clients need, and limited by address.
func (s *Server) serveRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && !s.registrations.allow(clientIP(r)) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Retry-After", "60")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"invalid_request","error_description":"too many registrations from this address: wait a minute"}` + "\n"))
		return
	}
	siteOf(r).cahut.OAuth.ServeRegister(w, r)
}
