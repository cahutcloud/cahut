package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"

	"cahut.cloud/internal/oauth"
)

// Each Cahut is the OAuth authorization server of its MCP endpoint. The
// servers of AI clients reach its metadata, token, registration and
// revocation endpoints on port 443; people's browsers reach its consent
// screen in the console. See docs/adr/0016-oauth-for-ai-clients.md and
// docs/adr/0017-oauth-endpoints-on-port-443.md.

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
	if r.Method == http.MethodPost && !s.registrations.allow(clientKey(r)) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Retry-After", "60")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"invalid_request","error_description":"too many registrations from this address: wait a minute"}` + "\n"))
		return
	}
	siteOf(r).cahut.OAuth.ServeRegister(w, r)
}

// serveUnknownMetadata answers the well-known addresses that Cahut does not
// serve, such as OpenID Connect discovery, with a 404 in JSON rather than a
// page: a client that probes them reads that there is nothing there.
func serveUnknownMetadata(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"error":"not_found","error_description":"Cahut serves OAuth 2.0 authorization server metadata at /.well-known/oauth-authorization-server, not OpenID Connect"}` + "\n"))
}

// logged logs each request to h, an endpoint of the OAuth protocol: what an
// AI client asked, and how Cahut answered. Clients run on their vendor's
// servers, out of the person's sight, so these lines are how an
// administrator follows a connection that fails. They hold no query, no
// token and no successful answer: only the error of a refused request.
func (s *Server) logged(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &recorder{ResponseWriter: w, status: http.StatusOK}
		h(rec, r)
		attrs := []any{"cahut", hostOf(r), "method", r.Method, "path", r.URL.Path, "status", rec.status,
			"client", clientIP(r), "user_agent", r.UserAgent(), "duration", time.Since(start).Round(time.Millisecond)}
		if rec.status >= 400 {
			var refusal struct {
				Error       string `json:"error"`
				Description string `json:"error_description"`
			}
			if json.Unmarshal(rec.body.Bytes(), &refusal) == nil && refusal.Error != "" {
				attrs = append(attrs, "error", refusal.Error, "description", refusal.Description)
			}
		}
		s.log.Info("oauth", attrs...)
	}
}

// recorder keeps the status of a response, and the start of its body when
// it is an error.
type recorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *recorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *recorder) Write(b []byte) (int, error) {
	if r.status >= 400 && r.body.Len() < 1024 {
		r.body.Write(b[:min(len(b), 1024-r.body.Len())])
	}
	return r.ResponseWriter.Write(b)
}

// Flush lets the MCP endpoint stream through a recorder.
func (r *recorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
