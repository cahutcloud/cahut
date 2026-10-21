// Package account manages the people of a Cahut and the personal tokens
// that their AIs present.
package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Register adds the account operations to r.
func Register(r *operation.Registry) {
	operation.Register(r, CreateUser)
	operation.Register(r, CreateToken)
	operation.Register(r, ListTokens)
	operation.Register(r, RevokeToken)
}

// User is a person of a Cahut.
type User struct {
	ID           int64  `json:"id"`
	Handle       string `json:"handle"`
	Name         string `json:"name"`
	Role         string `json:"role" jsonschema:"admin or member of the Cahut"`
	PersonalRoom string `json:"personal_room" jsonschema:"name of the person's personal Room, the handle"`
}

// CreateUserInput is the input of create_user.
type CreateUserInput struct {
	Handle string `json:"handle"`
	Name   string `json:"name"`
	Admin  bool   `json:"admin,omitempty"`
}

// A handle is also the name of the person's personal Room, so it is a valid
// Room name too.
var validHandle = regexp.MustCompile(`^[a-z][a-z0-9-]{0,30}[a-z0-9]$`)

// CreateUser creates a person and their personal Room, named after their
// handle.
var CreateUser = &operation.Op[CreateUserInput, User]{
	Name:        "create_user",
	Description: "Create a person, with their personal Room.",
	AdminOnly:   true,
	Run: func(ctx context.Context, c *operation.Call, in CreateUserInput) (User, error) {
		if !validHandle.MatchString(in.Handle) {
			return User{}, operation.Invalidf("handle %q: use 2 to 32 lowercase letters, digits and hyphens, "+
				"starting with a letter and ending with a letter or a digit", in.Handle)
		}
		if address.Reserved(in.Handle) {
			return User{}, operation.Invalidf("handle %q is reserved: it names a page of every Cahut", in.Handle)
		}
		name := strings.TrimSpace(in.Name)
		if name == "" {
			return User{}, operation.Invalidf("name is empty")
		}
		u := User{Handle: in.Handle, Name: name, Role: "member", PersonalRoom: in.Handle}
		if in.Admin {
			u.Role = "admin"
		}
		c.Log(u.Handle, u.Role)

		var taken bool
		err := c.Tx.QueryRowContext(ctx, `
			SELECT EXISTS (SELECT 1 FROM users WHERE handle = ?) OR EXISTS (SELECT 1 FROM rooms WHERE name = ?)`,
			u.Handle, u.Handle).Scan(&taken)
		if err != nil {
			return User{}, err
		}
		if taken {
			return User{}, operation.Invalidf("handle %q is taken, by a person or a Room", u.Handle)
		}
		now := store.FormatTime(c.Now)
		err = c.Tx.QueryRowContext(ctx, "INSERT INTO users (handle, name, role, created_at) VALUES (?, ?, ?, ?) RETURNING id",
			u.Handle, u.Name, u.Role, now).Scan(&u.ID)
		if err != nil {
			return User{}, err
		}
		var room int64
		err = c.Tx.QueryRowContext(ctx, "INSERT INTO rooms (name, kind, title, created_at) VALUES (?, 'personal', ?, ?) RETURNING id",
			u.PersonalRoom, u.Name, now).Scan(&room)
		if err != nil {
			return User{}, err
		}
		_, err = c.Tx.ExecContext(ctx, "INSERT INTO room_members (room_id, user_id, role) VALUES (?, ?, 'owner')",
			room, u.ID)
		return u, err
	},
}

// TokenPrefix starts every personal token, so that secret scanners and
// people can recognize one.
const TokenPrefix = "cahut_pat_" //nolint:gosec // a prefix, not a credential

// CreateTokenInput is the input of create_token.
type CreateTokenInput struct {
	User string `json:"user" jsonschema:"handle of the person the AI acts for"`
	Name string `json:"name" jsonschema:"what the token is for, such as the AI that uses it"`
}

// NewToken is a token just created. Token is shown once and never stored.
type NewToken struct {
	ID    int64  `json:"id"`
	User  string `json:"user"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

// CreateToken creates a personal token, which lets an AI act on behalf of
// a person.
var CreateToken = &operation.Op[CreateTokenInput, NewToken]{
	Name:        "create_token",
	Description: "Create a personal token that lets an AI act on behalf of a person.",
	AdminOnly:   true,
	Run: func(ctx context.Context, c *operation.Call, in CreateTokenInput) (NewToken, error) {
		name := strings.TrimSpace(in.Name)
		if name == "" {
			return NewToken{}, operation.Invalidf("name is empty")
		}
		var userID int64
		err := c.Tx.QueryRowContext(ctx, "SELECT id FROM users WHERE handle = ?", in.User).Scan(&userID)
		if errors.Is(err, sql.ErrNoRows) {
			return NewToken{}, operation.NotFoundf("no user %q", in.User)
		}
		if err != nil {
			return NewToken{}, err
		}
		token := TokenPrefix + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes(32)))
		t := NewToken{User: in.User, Name: name, Token: token}
		err = c.Tx.QueryRowContext(ctx, "INSERT INTO tokens (user_id, name, hash, created_at) VALUES (?, ?, ?, ?) RETURNING id",
			userID, name, hash(token), store.FormatTime(c.Now)).Scan(&t.ID)
		if err != nil {
			return NewToken{}, err
		}
		c.Log(fmt.Sprintf("token %d of %s", t.ID, in.User), name)
		return t, nil
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
	User string `json:"user,omitempty" jsonschema:"only the tokens of this person"`
}

// TokenList is the output of list_tokens.
type TokenList struct {
	Tokens []TokenInfo `json:"tokens"`
}

// ListTokens lists personal tokens, newest first.
var ListTokens = &operation.Op[ListTokensInput, TokenList]{
	Name:        "list_tokens",
	Description: "List personal tokens, newest first.",
	AdminOnly:   true,
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, in ListTokensInput) (TokenList, error) {
		rows, err := c.Tx.QueryContext(ctx, `
			SELECT t.id, u.handle, t.name, t.created_at, t.last_used_at, t.revoked_at
			FROM tokens t JOIN users u ON u.id = t.user_id
			WHERE ? = '' OR u.handle = ?
			ORDER BY t.id DESC`, in.User, in.User)
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
	Description: "Revoke a personal token. The AI that used it loses access at once.",
	AdminOnly:   true,
	Run: func(ctx context.Context, c *operation.Call, in RevokeTokenInput) (TokenInfo, error) {
		c.Log(fmt.Sprintf("token %d", in.ID), "")
		t := TokenInfo{ID: in.ID}
		var revoked sql.NullString
		err := c.Tx.QueryRowContext(ctx, `
			SELECT u.handle, t.name, t.revoked_at FROM tokens t JOIN users u ON u.id = t.user_id
			WHERE t.id = ?`, in.ID).Scan(&t.User, &t.Name, &revoked)
		if errors.Is(err, sql.ErrNoRows) {
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
	TokenID   int64
	TokenName string
	UserID    int64
	Handle    string
}

// ErrInvalidToken means the token is unknown or revoked.
var ErrInvalidToken = errors.New("invalid or revoked token")

// Verify checks a personal token and returns who it acts for. It records
// when the token was last used, at most once a minute.
func Verify(ctx context.Context, db *store.DB, token string, now time.Time) (Owner, error) {
	if !strings.HasPrefix(token, TokenPrefix) {
		return Owner{}, ErrInvalidToken
	}
	var o Owner
	err := db.Read.QueryRowContext(ctx, `
		SELECT t.id, t.name, u.id, u.handle FROM tokens t JOIN users u ON u.id = t.user_id
		WHERE t.hash = ? AND t.revoked_at IS NULL`, hash(token)).Scan(&o.TokenID, &o.TokenName, &o.UserID, &o.Handle)
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

func hash(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b) // crypto/rand.Read never fails
	return b
}

func parseNullTime(s sql.NullString) (*time.Time, error) {
	if !s.Valid {
		return nil, nil
	}
	t, err := store.ParseTime(s.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
