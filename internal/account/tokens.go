package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// TokenPrefix starts every personal token, so that secret scanners and
// people can recognize one.
const TokenPrefix = "cahut_pat_" //nolint:gosec // a prefix, not a credential

// CreateTokenInput is the input of create_token.
type CreateTokenInput struct {
	User string `json:"user,omitempty" jsonschema:"handle of the person the AI acts for; yourself when empty. Only administrators create tokens for others"`
	Name string `json:"name" jsonschema:"what the token is for, such as the AI that uses it"`
}

// NewToken is a token just created. Token is shown once and never stored.
type NewToken struct {
	ID    int64  `json:"id"`
	User  string `json:"user"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

// owner returns the person a token operation is about: the one named in
// user, or the caller when user is empty. Only administrators act on the
// tokens of others.
func owner(ctx context.Context, c *operation.Call, op, user string) (person, error) {
	if user == "" {
		if c.Actor.UserID == 0 {
			return person{}, operation.Invalidf("user is empty: name the person the AI acts for")
		}
		user = c.Actor.Handle
	}
	if user != c.Actor.Handle && !c.Admin {
		return person{}, operation.Forbiddenf("%s: only administrators of the Cahut act on the tokens of others", op)
	}
	return findUser(ctx, c.Tx, user)
}

// CreateToken creates a personal token, which lets an AI act on behalf of
// a person.
var CreateToken = &operation.Op[CreateTokenInput, NewToken]{
	Name:        "create_token",
	Title:       "Create a personal token",
	Description: "Create a personal token that lets an AI act on behalf of a person.",
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, in CreateTokenInput) (NewToken, error) {
		name := strings.TrimSpace(in.Name)
		if name == "" {
			return NewToken{}, operation.Invalidf("name is empty")
		}
		p, err := owner(ctx, c, "create_token", in.User)
		if err != nil {
			return NewToken{}, err
		}
		t := NewToken{User: p.Handle, Name: name, Token: credential.Token(TokenPrefix)}
		err = c.Tx.QueryRowContext(ctx, "INSERT INTO tokens (user_id, name, hash, created_at) VALUES (?, ?, ?, ?) RETURNING id",
			p.ID, name, credential.Hash(t.Token), store.FormatTime(c.Now)).Scan(&t.ID)
		if err != nil {
			return NewToken{}, err
		}
		c.Log(fmt.Sprintf("token %d of %s", t.ID, p.Handle), name)
		return t, mail.OutboxOf(c).Notify(ctx, c, p.ID, mail.Always, tokenLetter(name))
	},
}

// TokenInfo describes a token, without the token itself.
type TokenInfo struct {
	ID         int64      `json:"id"`
	User       string     `json:"user"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

// ListTokensInput is the input of list_tokens.
type ListTokensInput struct {
	User string `json:"user,omitempty" jsonschema:"only the tokens of this person; for a person who is not an administrator, their own tokens"`
}

// TokenList is the output of list_tokens.
type TokenList struct {
	Tokens []TokenInfo `json:"tokens"`
}

// ListTokens lists personal tokens, newest first: all of them for an
// administrator, their own for anyone else.
var ListTokens = &operation.Op[ListTokensInput, TokenList]{
	Name:        "list_tokens",
	Title:       "List personal tokens",
	Description: "List personal tokens, newest first.",
	PeopleOnly:  true,
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, in ListTokensInput) (TokenList, error) {
		user := in.User
		if !c.Admin {
			if user != "" && user != c.Actor.Handle {
				return TokenList{}, operation.Forbiddenf("list_tokens: only administrators of the Cahut see the tokens of others")
			}
			user = c.Actor.Handle
		}
		rows, err := c.Tx.QueryContext(ctx, `
			SELECT t.id, u.handle, t.name, t.created_at, t.last_used_at, t.revoked_at
			FROM tokens t JOIN users u ON u.id = t.user_id
			WHERE ? = '' OR u.handle = ?
			ORDER BY t.id DESC`, user, user)
		if err != nil {
			return TokenList{}, err
		}
		defer rows.Close()
		list := TokenList{Tokens: []TokenInfo{}}
		for rows.Next() {
			var (
				t               TokenInfo
				created         string
				lastUsed, revok sql.NullString
			)
			if err := rows.Scan(&t.ID, &t.User, &t.Name, &created, &lastUsed, &revok); err != nil {
				return TokenList{}, err
			}
			if t.CreatedAt, err = store.ParseTime(created); err != nil {
				return TokenList{}, err
			}
			if t.LastUsedAt, err = parseNullTime(lastUsed); err != nil {
				return TokenList{}, err
			}
			if t.RevokedAt, err = parseNullTime(revok); err != nil {
				return TokenList{}, err
			}
			list.Tokens = append(list.Tokens, t)
		}
		return list, rows.Err()
	},
}

// RevokeTokenInput is the input of revoke_token.
type RevokeTokenInput struct {
	ID int64 `json:"id"`
}

// RevokeToken revokes a personal token at once.
var RevokeToken = &operation.Op[RevokeTokenInput, TokenInfo]{
	Name:        "revoke_token",
	Title:       "Revoke a personal token",
	Description: "Revoke a personal token. The AI that used it loses access at once.",
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, in RevokeTokenInput) (TokenInfo, error) {
		c.Log(fmt.Sprintf("token %d", in.ID), "")
		t := TokenInfo{ID: in.ID}
		var userID int64
		var revoked sql.NullString
		err := c.Tx.QueryRowContext(ctx, `
			SELECT u.id, u.handle, t.name, t.revoked_at FROM tokens t JOIN users u ON u.id = t.user_id
			WHERE t.id = ?`, in.ID).Scan(&userID, &t.User, &t.Name, &revoked)
		// The tokens of others do not exist for those who may not see them.
		if errors.Is(err, sql.ErrNoRows) || err == nil && userID != c.Actor.UserID && !c.Admin {
			return TokenInfo{}, operation.NotFoundf("no token %d", in.ID)
		}
		if err != nil {
			return TokenInfo{}, err
		}
		c.Log(fmt.Sprintf("token %d of %s", in.ID, t.User), t.Name)
		if revoked.Valid {
			return TokenInfo{}, operation.Invalidf("token %d is already revoked", in.ID)
		}
		now := c.Now
		if _, err := c.Tx.ExecContext(ctx, "UPDATE tokens SET revoked_at = ? WHERE id = ?", store.FormatTime(now), in.ID); err != nil {
			return TokenInfo{}, err
		}
		t.RevokedAt = &now
		return t, nil
	},
}

// Owner is who a valid token acts for.
type Owner struct {
	TokenID    int64  // the personal token, or zero
	Connection int64  // the connection over OAuth, or zero
	TokenName  string // the name of the token, or of the connected client
	UserID     int64
	Handle     string
	Admin      bool // the person administers the Cahut

	// What the person granted the AI: a personal token reaches everything
	// its person may do.
	ReadOnly bool
	Rooms    []int64 // every Room of the person when nil
}

// ErrInvalidToken means the token is unknown or revoked, or the person it
// acts for is disabled.
var ErrInvalidToken = errors.New("invalid or revoked token")

// Verify checks a personal token and returns who it acts for. It records
// when the token was last used, at most once a minute. The tokens of a
// disabled person are suspended.
func Verify(ctx context.Context, db *store.DB, token string, now time.Time) (Owner, error) {
	if !strings.HasPrefix(token, TokenPrefix) {
		return Owner{}, ErrInvalidToken
	}
	var o Owner
	err := db.Read.QueryRowContext(ctx, `
		SELECT t.id, t.name, u.id, u.handle, u.role = 'admin' FROM tokens t JOIN users u ON u.id = t.user_id
		WHERE t.hash = ? AND t.revoked_at IS NULL AND u.disabled_at IS NULL`,
		credential.Hash(token)).Scan(&o.TokenID, &o.TokenName, &o.UserID, &o.Handle, &o.Admin)
	if errors.Is(err, sql.ErrNoRows) {
		return Owner{}, ErrInvalidToken
	}
	if err != nil {
		return Owner{}, err
	}
	_, err = db.Write.ExecContext(ctx, `
		UPDATE tokens SET last_used_at = ? WHERE id = ? AND (last_used_at IS NULL OR last_used_at < ?)`,
		store.FormatTime(now), o.TokenID, store.FormatTime(now.Add(-time.Minute)))
	return o, err
}
