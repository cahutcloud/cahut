package oauth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// MaxTokenRequestSize bounds the body of a request to the token and
// revocation endpoints.
const MaxTokenRequestSize = 64 << 10

// Tokens is the answer of the token endpoint.
type Tokens struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// errGrant is a refused token request, in the terms of RFC 6749.
type errGrant struct{ code, description string }

func (e *errGrant) Error() string { return e.code + ": " + e.description }

func invalidGrant(description string) error {
	return &errGrant{"invalid_grant", description}
}

// replay is a code or a refresh token presented again: its connection
// ends.
type replay struct {
	connection int64
	reason     string
}

func (r *replay) Error() string { return r.reason }

// ServeToken answers the token endpoint: it exchanges an authorization code,
// or a refresh token, for new tokens.
func (s *Server) ServeToken(w http.ResponseWriter, r *http.Request) {
	if preflight(w, r, "POST") {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxTokenRequestSize)
	if err := r.ParseForm(); err != nil {
		protocolError(w, http.StatusBadRequest, "invalid_request", "the body must be a form, application/x-www-form-urlencoded")
		return
	}
	form := r.PostForm
	client := form.Get("client_id")
	if id, _, ok := r.BasicAuth(); ok && id != "" {
		client = id // a public client may name itself this way, without a secret
	}
	if client == "" {
		client = assertedClient(form.Get("client_assertion"))
	}
	var tokens Tokens
	var err error
	switch form.Get("grant_type") {
	case "authorization_code":
		tokens, err = s.exchange(r.Context(), client, form.Get("code"), form.Get("redirect_uri"), form.Get("code_verifier"), form.Get("resource"))
	case "refresh_token":
		tokens, err = s.refresh(r.Context(), client, form.Get("refresh_token"), form.Get("resource"))
	case "":
		err = &errGrant{"invalid_request", "grant_type is missing"}
	default:
		err = &errGrant{"unsupported_grant_type", "Cahut takes authorization_code and refresh_token"}
	}
	var rp *replay
	if errors.As(err, &rp) {
		s.end(r.Context(), rp.connection, rp.reason)
		err = invalidGrant(rp.reason + ": the connection has ended, and the person must connect the AI again")
	}
	var refused *errGrant
	if errors.As(err, &refused) {
		protocolError(w, http.StatusBadRequest, refused.code, refused.description)
		return
	}
	if err != nil {
		protocolError(w, http.StatusInternalServerError, "server_error", "")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	writeJSON(w, http.StatusOK, tokens)
}

// assertedClient returns the client that a client assertion (RFC 7523)
// names, for a client that authenticates with a key and sends no client_id.
// Cahut takes public clients only and does not check the signature: the
// name serves to match the client of the code, which PKCE binds anyway.
func assertedClient(assertion string) string {
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Subject string `json:"sub"`
	}
	if json.Unmarshal(payload, &claims) != nil {
		return ""
	}
	return claims.Subject
}

var verifierFormat = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`)

// exchange exchanges an authorization code for tokens.
func (s *Server) exchange(ctx context.Context, client, code, redirect, verifier, resource string) (Tokens, error) {
	if code == "" {
		return Tokens{}, &errGrant{"invalid_request", "code is missing"}
	}
	tx, err := s.db.Write.BeginTx(ctx, nil)
	if err != nil {
		return Tokens{}, err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	now := s.Now().UTC().Truncate(time.Millisecond)
	var (
		connection                                     int64
		clientID, redirectURI, challenge, codeResource string
		expires                                        string
		used                                           sql.NullString
	)
	err = tx.QueryRowContext(ctx, `
		SELECT connection_id, client_id, redirect_uri, challenge, resource, expires_at, used_at
		FROM oauth_codes WHERE hash = ?`, credential.Hash(code)).
		Scan(&connection, &clientID, &redirectURI, &challenge, &codeResource, &expires, &used)
	if errors.Is(err, sql.ErrNoRows) {
		return Tokens{}, invalidGrant("unknown authorization code")
	}
	if err != nil {
		return Tokens{}, err
	}
	if used.Valid {
		return Tokens{}, &replay{connection, "an authorization code was used twice"}
	}
	expiresAt, err := store.ParseTime(expires)
	if err != nil {
		return Tokens{}, err
	}
	switch {
	case !now.Before(expiresAt):
		return Tokens{}, invalidGrant("the authorization code has expired: authorize again")
	case client != clientID:
		return Tokens{}, invalidGrant("the authorization code was issued to another client")
	case redirect != "" && redirect != redirectURI:
		return Tokens{}, invalidGrant("redirect_uri differs from the one of the authorization request")
	case resource != "" && !s.sameResource(resource):
		return Tokens{}, &errGrant{"invalid_target", "this server grants access to " + s.cfg.Resource + " only"}
	case !verifierFormat.MatchString(verifier):
		return Tokens{}, invalidGrant("code_verifier is missing or malformed")
	}
	sum := sha256.Sum256([]byte(verifier))
	if subtle.ConstantTimeCompare([]byte(base64.RawURLEncoding.EncodeToString(sum[:])), []byte(challenge)) != 1 {
		return Tokens{}, invalidGrant("code_verifier does not match the code_challenge")
	}
	if _, err := tx.ExecContext(ctx, "UPDATE oauth_codes SET used_at = ? WHERE hash = ?", store.FormatTime(now), credential.Hash(code)); err != nil {
		return Tokens{}, err
	}
	tokens, err := s.issue(ctx, tx, connection, now)
	if err != nil {
		return Tokens{}, err
	}
	return tokens, tx.Commit()
}

// refresh replaces a refresh token with new tokens.
func (s *Server) refresh(ctx context.Context, client, token, resource string) (Tokens, error) {
	if token == "" {
		return Tokens{}, &errGrant{"invalid_request", "refresh_token is missing"}
	}
	if resource != "" && !s.sameResource(resource) {
		return Tokens{}, &errGrant{"invalid_target", "this server grants access to " + s.cfg.Resource + " only"}
	}
	tx, err := s.db.Write.BeginTx(ctx, nil)
	if err != nil {
		return Tokens{}, err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	now := s.Now().UTC().Truncate(time.Millisecond)
	var (
		id, connection    int64
		clientID, expires string
		replaced          sql.NullString
	)
	err = tx.QueryRowContext(ctx, `
		SELECT t.id, t.connection_id, c.client_id, t.expires_at, t.replaced_at
		FROM oauth_tokens t JOIN connections c ON c.id = t.connection_id
		WHERE t.hash = ? AND t.kind = 'refresh'`, credential.Hash(token)).
		Scan(&id, &connection, &clientID, &expires, &replaced)
	if errors.Is(err, sql.ErrNoRows) {
		return Tokens{}, invalidGrant("unknown or revoked refresh token: the person must connect the AI again")
	}
	if err != nil {
		return Tokens{}, err
	}
	if client != "" && client != clientID {
		return Tokens{}, invalidGrant("the refresh token was issued to another client")
	}
	if replaced.Valid {
		at, err := store.ParseTime(replaced.String)
		if err != nil {
			return Tokens{}, err
		}
		if now.Sub(at) < RefreshGrace {
			return Tokens{}, invalidGrant("the refresh token was just replaced: use the new one")
		}
		return Tokens{}, &replay{connection, "a refresh token was used after it was replaced"}
	}
	expiresAt, err := store.ParseTime(expires)
	if err != nil {
		return Tokens{}, err
	}
	if !now.Before(expiresAt) {
		return Tokens{}, invalidGrant("the refresh token has expired: the person must connect the AI again")
	}
	if _, err := tx.ExecContext(ctx, "UPDATE oauth_tokens SET replaced_at = ? WHERE id = ?", store.FormatTime(now), id); err != nil {
		return Tokens{}, err
	}
	tokens, err := s.issue(ctx, tx, connection, now)
	if err != nil {
		return Tokens{}, err
	}
	return tokens, tx.Commit()
}

// issue issues an access token and a refresh token to the connection, if
// it has not ended and its person is not disabled. It deletes the tokens of
// the connection that have expired.
func (s *Server) issue(ctx context.Context, tx *sql.Tx, connection int64, now time.Time) (Tokens, error) {
	var scope string
	var ended, disabled sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT c.scope, c.ended_at, u.disabled_at FROM connections c JOIN users u ON u.id = c.user_id
		WHERE c.id = ?`, connection).Scan(&scope, &ended, &disabled)
	if err != nil {
		return Tokens{}, err
	}
	switch {
	case ended.Valid:
		return Tokens{}, invalidGrant("the connection has ended: the person must connect the AI again")
	case disabled.Valid:
		return Tokens{}, invalidGrant("the person this AI acts for is disabled")
	}
	// A replaced refresh token is kept until it expires, to recognize a
	// replay.
	if _, err := tx.ExecContext(ctx, "DELETE FROM oauth_tokens WHERE connection_id = ? AND expires_at <= ?",
		connection, store.FormatTime(now)); err != nil {
		return Tokens{}, err
	}
	t := Tokens{
		AccessToken:  credential.Token(AccessPrefix),
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessLifetime / time.Second),
		RefreshToken: credential.Token(RefreshPrefix),
		Scope:        scope,
	}
	for _, tok := range []struct {
		value, kind string
		lifetime    time.Duration
	}{{t.AccessToken, "access", AccessLifetime}, {t.RefreshToken, "refresh", RefreshLifetime}} {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO oauth_tokens (hash, connection_id, kind, created_at, expires_at) VALUES (?, ?, ?, ?, ?)`,
			credential.Hash(tok.value), connection, tok.kind, store.FormatTime(now), store.FormatTime(now.Add(tok.lifetime)))
		if err != nil {
			return Tokens{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, "UPDATE connections SET last_used_at = ? WHERE id = ?", store.FormatTime(now), connection); err != nil {
		return Tokens{}, err
	}
	return t, nil
}

// ServeRevoke answers the revocation endpoint (RFC 7009). Revoking a
// refresh token ends its connection; revoking an access token ends that
// token. It answers 200 whatever the token, as RFC 7009 asks.
func (s *Server) ServeRevoke(w http.ResponseWriter, r *http.Request) {
	if preflight(w, r, "POST") {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxTokenRequestSize)
	if err := r.ParseForm(); err != nil {
		protocolError(w, http.StatusBadRequest, "invalid_request", "the body must be a form, application/x-www-form-urlencoded")
		return
	}
	client := r.PostForm.Get("client_id")
	if id, _, ok := r.BasicAuth(); ok && id != "" {
		client = id
	}
	if err := s.revoke(r.Context(), client, r.PostForm.Get("token")); err != nil {
		protocolError(w, http.StatusServiceUnavailable, "temporarily_unavailable", "")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) revoke(ctx context.Context, client, token string) error {
	var id, connection int64
	var kind, clientID string
	err := s.db.Read.QueryRowContext(ctx, `
		SELECT t.id, t.connection_id, t.kind, c.client_id FROM oauth_tokens t JOIN connections c ON c.id = t.connection_id
		WHERE t.hash = ?`, credential.Hash(token)).Scan(&id, &connection, &kind, &clientID)
	if errors.Is(err, sql.ErrNoRows) || err == nil && client != "" && client != clientID {
		return nil
	}
	if err != nil {
		return err
	}
	if kind == "access" {
		_, err := s.db.Write.ExecContext(ctx, "DELETE FROM oauth_tokens WHERE id = ?", id)
		return err
	}
	s.end(ctx, connection, "revoked by its client")
	return nil
}

// end ends a connection that its client revoked, or whose code or refresh
// token came back, through the registry, so that the Logbook records it.
func (s *Server) end(ctx context.Context, connection int64, reason string) {
	var userID int64
	var handle, client string
	err := s.db.Read.QueryRowContext(ctx, `
		SELECT u.id, u.handle, c.client_name FROM connections c JOIN users u ON u.id = c.user_id WHERE c.id = ?`,
		connection).Scan(&userID, &handle, &client)
	if err != nil {
		return
	}
	actor := operation.Actor{Kind: operation.AI, Via: "oauth", UserID: userID, Handle: handle, Conn: connection, Client: client}
	_, _ = operation.Do(ctx, s.registry, endConnection, actor, endInput{Connection: connection, Reason: reason})
}

// AccessOwner returns who a valid OAuth access token acts for, and when the
// token expires. It records when the connection was last used, at most once
// a minute. The connections of a disabled person are suspended.
func (s *Server) AccessOwner(ctx context.Context, token string) (account.Owner, time.Time, error) {
	if !strings.HasPrefix(token, AccessPrefix) {
		return account.Owner{}, time.Time{}, account.ErrInvalidToken
	}
	now := s.Now().UTC()
	var o account.Owner
	var scope, expires string
	var all bool
	err := s.db.Read.QueryRowContext(ctx, `
		SELECT c.id, c.client_name, c.scope, c.all_rooms, t.expires_at, u.id, u.handle, u.role = 'admin'
		FROM oauth_tokens t JOIN connections c ON c.id = t.connection_id JOIN users u ON u.id = c.user_id
		WHERE t.hash = ? AND t.kind = 'access' AND t.expires_at > ? AND c.ended_at IS NULL AND u.disabled_at IS NULL`,
		credential.Hash(token), store.FormatTime(now)).
		Scan(&o.Connection, &o.TokenName, &scope, &all, &expires, &o.UserID, &o.Handle, &o.Admin)
	if errors.Is(err, sql.ErrNoRows) {
		return account.Owner{}, time.Time{}, account.ErrInvalidToken
	}
	if err != nil {
		return account.Owner{}, time.Time{}, err
	}
	expiresAt, err := store.ParseTime(expires)
	if err != nil {
		return account.Owner{}, time.Time{}, err
	}
	o.ReadOnly = scope == Read
	if !all {
		o.Rooms = []int64{}
		rows, err := s.db.Read.QueryContext(ctx, "SELECT room_id FROM connection_rooms WHERE connection_id = ?", o.Connection)
		if err != nil {
			return account.Owner{}, time.Time{}, err
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return account.Owner{}, time.Time{}, err
			}
			o.Rooms = append(o.Rooms, id)
		}
		if err := rows.Err(); err != nil {
			return account.Owner{}, time.Time{}, err
		}
	}
	_, err = s.db.Write.ExecContext(ctx, `
		UPDATE connections SET last_used_at = ? WHERE id = ? AND (last_used_at IS NULL OR last_used_at < ?)`,
		store.FormatTime(now), o.Connection, store.FormatTime(now.Add(-time.Minute)))
	if err != nil {
		return account.Owner{}, time.Time{}, fmt.Errorf("record the use of connection %d: %w", o.Connection, err)
	}
	return o, expiresAt, nil
}
