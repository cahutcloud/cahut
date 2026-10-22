package oauth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Client is an AI client, as it registered or as its metadata document
// describes it.
type Client struct {
	ID           string
	Name         string
	URI          string // its web page, as it gives it
	RedirectURIs []string
	// Document reports whether the client ID is the address of a Client ID
	// Metadata Document, whose host vouches for the client.
	Document bool
}

// client returns the client with the ID id: fetched from its metadata
// document when id is an https address, or read from those that
// registered. It returns an error wrapping operation.ErrInvalid when there
// is no such client.
func (s *Server) client(ctx context.Context, tx *sql.Tx, id string) (Client, error) {
	if id == "" {
		return Client{}, operation.Invalidf("the request names no client: client_id is empty")
	}
	if isDocumentID(id) {
		return s.docs.get(ctx, id)
	}
	var c Client
	var uri sql.NullString
	var redirects string
	err := tx.QueryRowContext(ctx, "SELECT id, name, uri, redirect_uris FROM oauth_clients WHERE id = ?", id).
		Scan(&c.ID, &c.Name, &uri, &redirects)
	if errors.Is(err, sql.ErrNoRows) {
		return Client{}, operation.Invalidf("no client %q is registered here: the AI should register again", id)
	}
	if err != nil {
		return Client{}, err
	}
	c.URI = uri.String
	if err := json.Unmarshal([]byte(redirects), &c.RedirectURIs); err != nil {
		return Client{}, fmt.Errorf("redirect URIs of client %s: %w", id, err)
	}
	return c, nil
}

// isDocumentID reports whether a client ID is the address of a Client ID
// Metadata Document: an https address with a path.
func isDocumentID(id string) bool {
	u, err := url.Parse(id)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.Path != "" && u.Path != "/" &&
		u.User == nil && u.Fragment == "" && u.RawFragment == ""
}

// redirectURI returns the redirect URI of a request: the one it names, when
// the client registered it, or the only one the client has. A loopback
// address in http matches whatever its port (RFC 8252).
func (c Client) redirectURI(requested string) (string, error) {
	if requested == "" {
		if len(c.RedirectURIs) == 1 {
			return c.RedirectURIs[0], nil
		}
		return "", operation.Invalidf("the request names no redirect_uri, and the client %s has several", c.Name)
	}
	for _, registered := range c.RedirectURIs {
		if requested == registered || loopbackMatch(registered, requested) {
			return requested, nil
		}
	}
	return "", operation.Invalidf("the redirect_uri %s is not one of the client %s: the AI should register again", requested, c.Name)
}

// loopbackMatch reports whether requested is registered with another port,
// on a loopback address in http.
func loopbackMatch(registered, requested string) bool {
	a, err := url.Parse(registered)
	if err != nil || a.Scheme != "http" || !isLoopback(a.Hostname()) {
		return false
	}
	b, err := url.Parse(requested)
	if err != nil || b.Scheme != "http" || b.Hostname() != a.Hostname() {
		return false
	}
	return a.Path == b.Path && a.RawQuery == b.RawQuery && b.User == nil && b.Fragment == ""
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip, err := netip.ParseAddr(host)
	return err == nil && ip.IsLoopback()
}

// refusedSchemes cannot be redirect URIs: a browser would run them, or
// read local files.
var refusedSchemes = []string{"javascript", "data", "file", "blob", "about", "vbscript", "filesystem"}

// checkRedirectURI checks a redirect URI that a client registers: https,
// http on a loopback address, or the private scheme of an application,
// without a fragment.
func checkRedirectURI(uri string) error {
	u, err := url.Parse(uri)
	switch {
	case err != nil || u.Scheme == "":
		return fmt.Errorf("%q is not an absolute address", uri)
	case len(uri) > 2000:
		return fmt.Errorf("%q is longer than 2000 bytes", uri)
	case u.Fragment != "" || strings.Contains(uri, "#"):
		return fmt.Errorf("%q has a fragment", uri)
	case slices.Contains(refusedSchemes, strings.ToLower(u.Scheme)):
		return fmt.Errorf("%q: the scheme %s cannot receive a code", uri, u.Scheme)
	case u.Scheme == "https" && u.Host == "":
		return fmt.Errorf("%q has no host", uri)
	case u.Scheme == "http" && !isLoopback(u.Hostname()):
		return fmt.Errorf("%q: http is for loopback addresses only, use https", uri)
	}
	return nil
}

// RegistrationRequest is what a client sends to register (RFC 7591).
// Unknown fields are ignored.
type RegistrationRequest struct {
	RedirectURIs            []string `json:"redirect_uris"`
	ClientName              string   `json:"client_name"`
	ClientURI               string   `json:"client_uri"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

// Registration is the answer to a registration.
type Registration struct {
	ClientID                string   `json:"client_id"`
	ClientIDIssuedAt        int64    `json:"client_id_issued_at"`
	ClientName              string   `json:"client_name"`
	ClientURI               string   `json:"client_uri,omitempty"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

// errClientMetadata is a registration that cannot be accepted.
type errClientMetadata struct{ code, description string }

func (e *errClientMetadata) Error() string { return e.code + ": " + e.description }

// register registers a public client, and deletes the clients that waited
// too long for a connection.
func (s *Server) register(ctx context.Context, req RegistrationRequest) (Registration, error) {
	if len(req.RedirectURIs) == 0 {
		return Registration{}, &errClientMetadata{"invalid_redirect_uri", "redirect_uris is empty"}
	}
	if len(req.RedirectURIs) > 20 {
		return Registration{}, &errClientMetadata{"invalid_redirect_uri", "a client has at most 20 redirect URIs"}
	}
	for _, uri := range req.RedirectURIs {
		if err := checkRedirectURI(uri); err != nil {
			return Registration{}, &errClientMetadata{"invalid_redirect_uri", err.Error()}
		}
	}
	// Clients list the grants they can use, such as jwt-bearer for claude.ai,
	// and keep those the answer names (RFC 7591, section 3.2.1): Cahut
	// refuses a client only when it cannot use the authorization code.
	if len(req.GrantTypes) > 0 && !slices.Contains(req.GrantTypes, "authorization_code") {
		return Registration{}, &errClientMetadata{"invalid_client_metadata", "Cahut grants authorization codes only: grant_types must include authorization_code"}
	}
	if len(req.ResponseTypes) > 0 && !slices.Contains(req.ResponseTypes, "code") {
		return Registration{}, &errClientMetadata{"invalid_client_metadata", "Cahut answers response_type=code only: response_types must include code"}
	}
	name := strings.TrimSpace(req.ClientName)
	if name == "" {
		u, _ := url.Parse(req.RedirectURIs[0])
		name = u.Host
		if name == "" {
			name = "An AI client"
		}
	}
	if len(name) > 200 {
		name = name[:200]
	}
	uri := ""
	if u, err := url.Parse(req.ClientURI); err == nil && u.Scheme == "https" && u.Host != "" && len(req.ClientURI) <= 2000 {
		uri = req.ClientURI
	}
	redirects, err := json.Marshal(req.RedirectURIs)
	if err != nil {
		return Registration{}, err
	}
	now := s.Now().UTC()
	reg := Registration{
		ClientID:                credential.Token("cahut_cli_"),
		ClientIDIssuedAt:        now.Unix(),
		ClientName:              name,
		ClientURI:               uri,
		RedirectURIs:            req.RedirectURIs,
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "none",
	}
	tx, err := s.db.Write.BeginTx(ctx, nil)
	if err != nil {
		return Registration{}, err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	if _, err := tx.ExecContext(ctx, "DELETE FROM oauth_clients WHERE connected_at IS NULL AND created_at < ?",
		store.FormatTime(now.Add(-ClientLifetime))); err != nil {
		return Registration{}, err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO oauth_clients (id, name, uri, redirect_uris, created_at) VALUES (?, ?, ?, ?, ?)",
		reg.ClientID, name, sql.NullString{String: uri, Valid: uri != ""}, string(redirects), store.FormatTime(now))
	if err != nil {
		return Registration{}, err
	}
	return reg, tx.Commit()
}

// MaxRegistrationSize bounds the body of a registration.
const MaxRegistrationSize = 64 << 10

// ServeRegister answers a dynamic registration (RFC 7591).
func (s *Server) ServeRegister(w http.ResponseWriter, r *http.Request) {
	if preflight(w, r, "POST") {
		return
	}
	var req RegistrationRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxRegistrationSize)).Decode(&req); err != nil {
		protocolError(w, http.StatusBadRequest, "invalid_client_metadata", "the body is not a JSON object of client metadata")
		return
	}
	reg, err := s.register(r.Context(), req)
	var refused *errClientMetadata
	if errors.As(err, &refused) {
		protocolError(w, http.StatusBadRequest, refused.code, refused.description)
		return
	}
	if err != nil {
		protocolError(w, http.StatusInternalServerError, "server_error", "")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, reg)
}

// Client ID Metadata Documents are fetched within these bounds.
const (
	documentTimeout  = 5 * time.Second
	documentMaxSize  = 16 << 10
	documentLifetime = time.Hour
	documentsKept    = 1000
)

// documents fetches Client ID Metadata Documents and keeps them a while.
type documents struct {
	client *http.Client

	mu   sync.Mutex
	kept map[string]document
}

type document struct {
	client  Client
	fetched time.Time
}

func newDocuments() *documents {
	return &documents{client: guardedClient(), kept: map[string]document{}}
}

// guardedClient fetches public addresses only. It checks the address it
// connects to, after the name is resolved, so that a name that resolves
// first to a public address and then to a private one gains nothing. It
// follows no redirect and uses no proxy.
func guardedClient() *http.Client {
	dialer := &net.Dialer{
		Timeout: documentTimeout,
		Control: func(_, address string, _ syscall.RawConn) error {
			ap, err := netip.ParseAddrPort(address)
			if err != nil {
				return err
			}
			if !public(ap.Addr()) {
				return fmt.Errorf("%s is not a public address", ap.Addr())
			}
			return nil
		},
	}
	return &http.Client{
		Timeout: documentTimeout,
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			TLSHandshakeTimeout:   documentTimeout,
			ResponseHeaderTimeout: documentTimeout,
			MaxIdleConns:          10,
			IdleConnTimeout:       time.Minute,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New("a client metadata document does not redirect")
		},
	}
}

// nonPublic holds the ranges that netip does not classify, and that a
// server must not be made to reach.
var nonPublic = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),       // this network
	netip.MustParsePrefix("100.64.0.0/10"),   // shared address space, carrier-grade NAT
	netip.MustParsePrefix("192.0.0.0/24"),    // IETF protocol assignments
	netip.MustParsePrefix("198.18.0.0/15"),   // benchmarking
	netip.MustParsePrefix("240.0.0.0/4"),     // reserved, and broadcast
	netip.MustParsePrefix("64:ff9b::/96"),    // NAT64, which reaches IPv4 addresses
	netip.MustParsePrefix("64:ff9b:1::/48"),  // local NAT64
	netip.MustParsePrefix("2001:db8::/32"),   // documentation
	netip.MustParsePrefix("2002::/16"),       // 6to4, which reaches IPv4 addresses
	netip.MustParsePrefix("2001::/32"),       // Teredo
	netip.MustParsePrefix("fec0::/10"),       // site-local, deprecated
	netip.MustParsePrefix("100::/64"),        // discard
	netip.MustParsePrefix("192.88.99.0/24"),  // 6to4 relays
	netip.MustParsePrefix("198.51.100.0/24"), // documentation
	netip.MustParsePrefix("203.0.113.0/24"),  // documentation
	netip.MustParsePrefix("192.0.2.0/24"),    // documentation
}

// public reports whether ip is an address on the internet, as opposed to
// this machine, its networks or a reserved range.
func public(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, p := range nonPublic {
		if p.Contains(ip) {
			return false
		}
	}
	return true
}

// metadataDocument is a Client ID Metadata Document, as far as Cahut reads
// it.
type metadataDocument struct {
	ClientID                string   `json:"client_id"`
	ClientName              string   `json:"client_name"`
	ClientURI               string   `json:"client_uri"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	// TokenEndpointAuthMethods lists the methods the client can use, when
	// it can use several, as ChatGPT does: it prefers private_key_jwt and
	// can do without.
	TokenEndpointAuthMethods []string `json:"token_endpoint_auth_methods_supported"`
	ClientSecret             string   `json:"client_secret"`
}

// public reports whether the client of doc can authenticate as a public
// client, with the method none, which is the only one Cahut takes.
func (doc metadataDocument) public() bool {
	return doc.TokenEndpointAuthMethod == "" || doc.TokenEndpointAuthMethod == "none" ||
		slices.Contains(doc.TokenEndpointAuthMethods, "none")
}

// get returns the client that the metadata document at id describes.
func (d *documents) get(ctx context.Context, id string) (Client, error) {
	d.mu.Lock()
	kept, ok := d.kept[id]
	d.mu.Unlock()
	if ok && time.Since(kept.fetched) < documentLifetime {
		return kept.client, nil
	}
	c, err := d.fetch(ctx, id)
	if err != nil {
		return Client{}, operation.Invalidf("the client metadata document at %s: %v", id, err)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.kept) >= documentsKept {
		clear(d.kept)
	}
	d.kept[id] = document{client: c, fetched: time.Now()}
	return c, nil
}

func (d *documents) fetch(ctx context.Context, id string) (Client, error) {
	ctx, cancel := context.WithTimeout(ctx, documentTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, id, nil)
	if err != nil {
		return Client{}, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return Client{}, errors.New("it could not be fetched")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Client{}, fmt.Errorf("it answered %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, documentMaxSize+1))
	if err != nil {
		return Client{}, errors.New("it could not be read")
	}
	if len(body) > documentMaxSize {
		return Client{}, fmt.Errorf("it is larger than %d KiB", documentMaxSize>>10)
	}
	var doc metadataDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return Client{}, errors.New("it is not a JSON object")
	}
	switch {
	case doc.ClientID != id:
		return Client{}, fmt.Errorf("it names the client %q instead", doc.ClientID)
	case doc.ClientSecret != "":
		return Client{}, errors.New("it holds a client secret, which a public document must not")
	case !doc.public():
		return Client{}, fmt.Errorf("its client authenticates with %s only, and Cahut takes public clients, "+
			"which authenticate with none", doc.TokenEndpointAuthMethod)
	case len(doc.RedirectURIs) == 0:
		return Client{}, errors.New("it names no redirect URI")
	}
	for _, uri := range doc.RedirectURIs {
		if err := checkRedirectURI(uri); err != nil {
			return Client{}, err
		}
	}
	u, _ := url.Parse(id)
	c := Client{ID: id, Name: strings.TrimSpace(doc.ClientName), RedirectURIs: doc.RedirectURIs, Document: true}
	if c.Name == "" || len(c.Name) > 200 {
		c.Name = u.Host
	}
	if cu, err := url.Parse(doc.ClientURI); err == nil && cu.Scheme == "https" {
		c.URI = doc.ClientURI
	}
	return c, nil
}
