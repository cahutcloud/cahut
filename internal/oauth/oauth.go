// Package oauth is the authorization server of a Cahut, for AI clients that
// connect over MCP: its metadata, client registration, consent, tokens and
// revocation. A person grants an AI a scope and some Rooms; the grant is a
// connection, which the console calls a connected AI. See
// docs/adr/0006-in-house-oauth-server.md and
// docs/adr/0016-oauth-for-ai-clients.md.
package oauth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Prefixes of the tokens, so that secret scanners and people can recognize
// one.
const (
	AccessPrefix  = "cahut_oat_"
	RefreshPrefix = "cahut_ort_"
)

// Lifetimes. See docs/adr/0016-oauth-for-ai-clients.md.
const (
	CodeLifetime    = 5 * time.Minute
	AccessLifetime  = time.Hour
	RefreshLifetime = 30 * 24 * time.Hour
	// RefreshGrace is how long a replaced refresh token is refused without
	// ending its connection, for two refreshes sent at once.
	RefreshGrace = 30 * time.Second
	// ClientLifetime is how long a registered client waits for a first
	// connection before it is deleted.
	ClientLifetime = 7 * 24 * time.Hour
)

// Scopes: read allows the operations that change nothing, write everything
// the person may do.
const (
	Read  = "read"
	Write = "write"
)

// Config says where the authorization server and the resource it protects
// are.
type Config struct {
	// Issuer is the address of the authorization server, without a final
	// slash, such as https://cahut.example.com:8443.
	Issuer string
	// Resource is the address of the MCP endpoint, such as
	// https://cahut.example.com/mcp.
	Resource string
}

// Server is the authorization server of a Cahut.
type Server struct {
	db       *store.DB
	registry *operation.Registry
	cfg      Config
	docs     *documents

	// Now returns the current time. Tests replace it.
	Now func() time.Time
}

// New returns the authorization server of the Cahut whose database is db
// and whose registry is r. Register adds its operations to r.
func New(db *store.DB, r *operation.Registry, cfg Config) *Server {
	return &Server{db: db, registry: r, cfg: cfg, docs: newDocuments(), Now: time.Now}
}

// Register adds the operations of consent and of connected AIs to r.
func (s *Server) Register(r *operation.Registry) {
	operation.Register(r, s.describeOp())
	operation.Register(r, s.approveOp())
	operation.Register(r, s.denyOp())
	operation.Register(r, ListConnectedAIs)
	operation.Register(r, DisconnectAI)
}

// Issuer returns the address of the authorization server.
func (s *Server) Issuer() string { return s.cfg.Issuer }

// Resource returns the address of the resource the server protects.
func (s *Server) Resource() string { return s.cfg.Resource }

// Metadata is the authorization server's metadata (RFC 8414).
type Metadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	RegistrationEndpoint              string   `json:"registration_endpoint"`
	RevocationEndpoint                string   `json:"revocation_endpoint"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ResponseModesSupported            []string `json:"response_modes_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	RevocationAuthMethodsSupported    []string `json:"revocation_endpoint_auth_methods_supported"`
	ClientIDMetadataDocumentSupported bool     `json:"client_id_metadata_document_supported"`
	IssParameterSupported             bool     `json:"authorization_response_iss_parameter_supported"`
	ServiceDocumentation              string   `json:"service_documentation"`
}

// Metadata describes the authorization server.
func (s *Server) Metadata() Metadata {
	return Metadata{
		Issuer:                            s.cfg.Issuer,
		AuthorizationEndpoint:             s.cfg.Issuer + "/oauth/authorize",
		TokenEndpoint:                     s.cfg.Issuer + "/oauth/token",
		RegistrationEndpoint:              s.cfg.Issuer + "/oauth/register",
		RevocationEndpoint:                s.cfg.Issuer + "/oauth/revoke",
		ScopesSupported:                   []string{Read, Write},
		ResponseTypesSupported:            []string{"code"},
		ResponseModesSupported:            []string{"query"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		TokenEndpointAuthMethodsSupported: []string{"none"},
		RevocationAuthMethodsSupported:    []string{"none"},
		ClientIDMetadataDocumentSupported: true,
		IssParameterSupported:             true,
		ServiceDocumentation:              "https://docs.cahut.cloud/connect/ai-clients/",
	}
}

// ServeMetadata answers the metadata of the authorization server.
func (s *Server) ServeMetadata(w http.ResponseWriter, r *http.Request) {
	if preflight(w, r, "GET") {
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	writeJSON(w, http.StatusOK, s.Metadata())
}

// preflight sets the CORS headers of the protocol's endpoints, which read no
// cookie and no session, and answers a preflight request. It reports
// whether it answered.
func preflight(w http.ResponseWriter, r *http.Request, method string) bool {
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", method+", OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, MCP-Protocol-Version")
	h.Set("Access-Control-Max-Age", "86400")
	h.Set("X-Content-Type-Options", "nosniff")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	if r.Method != method {
		h.Set("Allow", method+", OPTIONS")
		protocolError(w, http.StatusMethodNotAllowed, "invalid_request", "this endpoint takes "+method+" requests")
		return true
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// protocolError answers an error of the token, registration or revocation
// endpoint, in the shape of RFC 6749.
func protocolError(w http.ResponseWriter, status int, code, description string) {
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, status, struct {
		Error       string `json:"error"`
		Description string `json:"error_description,omitempty"`
	}{code, description})
}

// sameResource reports whether resource names the resource of the server.
// Clients differ on a final slash and on the case of the host.
func (s *Server) sameResource(resource string) bool {
	return canonical(resource) == canonical(s.cfg.Resource)
}

func canonical(resource string) string {
	u, err := url.Parse(resource)
	if err != nil {
		return resource
	}
	u.Scheme, u.Host = strings.ToLower(u.Scheme), strings.ToLower(u.Host)
	u.Path = strings.TrimSuffix(u.Path, "/")
	return u.String()
}
