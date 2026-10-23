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
	"cahut.cloud/internal/version"
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
func (s *Server) client(ctx context.Context, c *operation.Call, id string) (Client, error) {
	if id == "" {
		return Client{}, operation.Invalidf("the request names no client: client_id is empty")
	}
	tx := c.Tx
	if isDocumentID(id) {
		return s.document(ctx, c, id)
	}
	var cl Client
	var uri sql.NullString
	var redirects string
	err := tx.QueryRowContext(ctx, "SELECT id, name, uri, redirect_uris FROM oauth_clients WHERE id = ?", id).
		Scan(&cl.ID, &cl.Name, &uri, &redirects)
	if errors.Is(err, sql.ErrNoRows) {
		return Client{}, operation.Invalidf("no client %q is registered here: the AI should register again", id)
	}
	if err != nil {
		return Client{}, err
	}
	cl.URI = uri.String
	if err := json.Unmarshal([]byte(redirects), &cl.RedirectURIs); err != nil {
		return Client{}, fmt.Errorf("redirect URIs of client %s: %w", id, err)
	}
	return cl, nil
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

// maxRedirectURI bounds a redirect URI, and maxRedirectURIs their number
// for a client that registers, so that registrations stay small.
const (
	maxRedirectURI  = 512
	maxRedirectURIs = 10
	// maxPendingClients bounds the clients that registered and never
	// connected: the oldest give way to new ones.
	maxPendingClients = 1000
)

func ascii(s string) bool {
	for i := range len(s) {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
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
	case len(uri) > maxRedirectURI:
		return fmt.Errorf("a redirect URI is longer than %d bytes", maxRedirectURI)
	case u.Fragment != "" || strings.Contains(uri, "#"):
		return fmt.Errorf("%q has a fragment", uri)
	case slices.Contains(refusedSchemes, strings.ToLower(u.Scheme)):
		return fmt.Errorf("%q: the scheme %s cannot receive a code", uri, u.Scheme)
	case u.Scheme == "https" && u.Host == "":
		return fmt.Errorf("%q has no host", uri)
	case !ascii(u.Host):
		// The consent screen shows the host: a letter of another alphabet
		// that looks like a Latin one would pass for another site.
		return fmt.Errorf("%q: write its host in ASCII, as its xn-- form for an international name", uri)
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
	if len(req.RedirectURIs) > maxRedirectURIs {
		return Registration{}, &errClientMetadata{"invalid_redirect_uri", fmt.Sprintf("a client has at most %d redirect URIs", maxRedirectURIs)}
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
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM oauth_clients WHERE id IN (
			SELECT id FROM oauth_clients WHERE connected_at IS NULL ORDER BY created_at DESC LIMIT -1 OFFSET ?)`,
		maxPendingClients-1); err != nil {
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
	documentsFetched = 8 // at once, on the whole server
	// documentStale is how long the last valid copy of a document stands
	// in for a document that no longer answers.
	documentStale = 30 * 24 * time.Hour
	// documentRetry is how long a copy that stood in is used before Cahut
	// tries the document again.
	documentRetry = 5 * time.Minute
)

// documents fetches Client ID Metadata Documents and keeps them in memory
// a while.
type documents struct {
	client *http.Client
	// fetching bounds the documents fetched at once.
	fetching chan struct{}

	mu   sync.Mutex
	docs map[string]document
}

type document struct {
	client  Client
	fetched time.Time
}

func newDocuments() *documents {
	return &documents{client: guardedClient(), fetching: make(chan struct{}, documentsFetched), docs: map[string]document{}}
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
	netip.MustParsePrefix("::/96"),           // IPv4-compatible, deprecated, such as ::127.0.0.1
	netip.MustParsePrefix("::ffff:0:0:0/96"), // IPv4-translated, such as ::ffff:0:127.0.0.1
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

// document returns the client that the metadata document at id describes.
// A valid document that answers is kept in memory a while, and its copy in
// the database once the call succeeds. When the document no longer answers,
// as when its host is down, its last valid copy stands in, for 30 days. A
// document that answers but is invalid is refused, whatever its copy says.
//
// The document is fetched before the call's transaction opens, by
// prefetch, so that a slow document holds no connection to the database:
// document itself never fetches.
func (s *Server) document(ctx context.Context, c *operation.Call, id string) (Client, error) {
	now := s.Now()
	if cl, ok := s.docs.kept(id, now); ok {
		return cl, nil
	}
	err, _ := c.Kept().(error)
	if err == nil {
		err = errors.New("it could not be fetched")
	}
	var stored, fetched string
	serr := c.Tx.QueryRowContext(ctx, "SELECT document, fetched_at FROM client_documents WHERE client_id = ? AND fetched_at > ?",
		id, store.FormatTime(c.Now.Add(-documentStale))).Scan(&stored, &fetched)
	if serr != nil {
		return Client{}, operation.Invalidf("the client metadata document at %s: %v", id, err)
	}
	cl, perr := parseDocument(id, []byte(stored))
	if perr != nil {
		return Client{}, operation.Invalidf("the client metadata document at %s: %v", id, err)
	}
	// Try the document again in a few minutes, not at every request.
	s.docs.keep(id, cl, now.Add(documentRetry-documentLifetime))
	return cl, nil
}

// prefetch fetches the metadata document of the client id, when id is the
// address of one that is not kept, and keeps its client: it runs before
// the transaction of the call, in Prepare. A document that cannot be
// fetched keeps its error in c, for document to say; one that answers but
// is invalid fails the call.
func (s *Server) prefetch(ctx context.Context, c *operation.Call, id string) error {
	now := s.Now()
	if !isDocumentID(id) {
		return nil
	}
	if _, ok := s.docs.kept(id, now); ok {
		return nil
	}
	body, err := s.docs.fetch(ctx, id)
	if err != nil {
		c.Keep(err)
		return nil
	}
	cl, err := parseDocument(id, body)
	if err != nil {
		return operation.Invalidf("the client metadata document at %s: %v", id, err)
	}
	s.docs.keep(id, cl, now)
	s.saveDocument(id, body, c.Now)
	return nil
}

// saveDocument keeps the copy of a valid document. It runs before the call
// that fetched it, outside its transaction: a copy that cannot be written
// waits for the next fetch.
func (s *Server) saveDocument(id string, body []byte, at time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), documentTimeout)
	defer cancel()
	_, _ = s.db.Write.ExecContext(ctx, `
		INSERT INTO client_documents (client_id, document, fetched_at) VALUES (?, ?, ?)
		ON CONFLICT (client_id) DO UPDATE SET document = excluded.document, fetched_at = excluded.fetched_at`,
		id, string(body), store.FormatTime(at))
}

// kept returns the client of the document at id, when it was kept less
// than documentLifetime before now.
func (d *documents) kept(id string, now time.Time) (Client, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	k, ok := d.docs[id]
	if !ok || now.Sub(k.fetched) >= documentLifetime {
		return Client{}, false
	}
	return k.client, true
}

// keep keeps the client of the document at id, as fetched at.
func (d *documents) keep(id string, c Client, at time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.docs) >= documentsKept {
		clear(d.docs)
	}
	d.docs[id] = document{client: c, fetched: at}
}

// fetch fetches the metadata document at id.
func (d *documents) fetch(ctx context.Context, id string) ([]byte, error) {
	select {
	case d.fetching <- struct{}{}:
		defer func() { <-d.fetching }()
	default:
		return nil, errors.New("too many documents are being fetched: try again in a few seconds")
	}
	ctx, cancel := context.WithTimeout(ctx, documentTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, id, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Cahut/"+version.Get().Version+" (+https://cahut.cloud)")
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, errors.New("it could not be fetched")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("it answered %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, documentMaxSize+1))
	if err != nil {
		return nil, errors.New("it could not be read")
	}
	if len(body) > documentMaxSize {
		return nil, fmt.Errorf("it is larger than %d KiB", documentMaxSize>>10)
	}
	return body, nil
}

// parseDocument checks the metadata document at id, and returns the client
// it describes.
func parseDocument(id string, body []byte) (Client, error) {
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
