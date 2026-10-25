package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Service tokens let a program on the server manage the people of the
// Cahut, through the admin API: the site where people sign up, for
// instance. See
// docs/adr/0035-a-program-on-the-server-manages-people-with-a-service-token.md.

// ServiceTokenPrefix starts every service token, so that secret scanners
// and people can recognize one.
const ServiceTokenPrefix = "cahut_svc_" //nolint:gosec // a prefix, not a credential

// CreateServiceTokenInput is the input of create_service_token.
type CreateServiceTokenInput struct {
	Name string `json:"name" jsonschema:"what uses the token, such as the site where people sign up"`
}

// NewServiceToken is a service token just created. Token is shown once and
// never stored.
type NewServiceToken struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

// CreateServiceToken creates a service token, for a program on the server.
var CreateServiceToken = &operation.Op[CreateServiceTokenInput, NewServiceToken]{
	Name:  "create_service_token",
	Title: "Create a service token",
	Description: "Create a service token, with which a program on this server manages the people of the Cahut " +
		"through the admin API, without a Knock: it creates them, gives them their activation link, and disables them.",
	Level:      operation.ServerAdmin,
	PeopleOnly: true,
	Run: func(ctx context.Context, c *operation.Call, in CreateServiceTokenInput) (NewServiceToken, error) {
		name := strings.TrimSpace(in.Name)
		if name == "" {
			return NewServiceToken{}, operation.Invalidf("name is empty: say what uses the token")
		}
		t := NewServiceToken{Name: name, Token: credential.Token(ServiceTokenPrefix)}
		err := c.Tx.QueryRowContext(ctx, "INSERT INTO service_tokens (name, hash, created_at) VALUES (?, ?, ?) RETURNING id",
			name, credential.Hash(t.Token), store.FormatTime(c.Now)).Scan(&t.ID)
		if err != nil {
			return NewServiceToken{}, err
		}
		c.Log(fmt.Sprintf("service token %d", t.ID), name)
		return t, nil
	},
}

// ServiceTokenInfo describes a service token, without the token itself.
type ServiceTokenInfo struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

// ListServiceTokensInput is the input of list_service_tokens.
type ListServiceTokensInput struct{}

// ServiceTokenList is the output of list_service_tokens.
type ServiceTokenList struct {
	Tokens []ServiceTokenInfo `json:"tokens"`
}

// ListServiceTokens lists the service tokens of the Cahut, newest first.
var ListServiceTokens = &operation.Op[ListServiceTokensInput, ServiceTokenList]{
	Name:        "list_service_tokens",
	Title:       "List service tokens",
	Description: "List the service tokens of the Cahut, newest first, with their last use.",
	Level:       operation.ServerAdmin,
	PeopleOnly:  true,
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, _ ListServiceTokensInput) (ServiceTokenList, error) {
		rows, err := c.Tx.QueryContext(ctx, "SELECT id, name, created_at, last_used_at, revoked_at FROM service_tokens ORDER BY id DESC")
		if err != nil {
			return ServiceTokenList{}, err
		}
		defer rows.Close()
		list := ServiceTokenList{Tokens: []ServiceTokenInfo{}}
		for rows.Next() {
			var (
				t               ServiceTokenInfo
				created         string
				lastUsed, revok sql.NullString
			)
			if err := rows.Scan(&t.ID, &t.Name, &created, &lastUsed, &revok); err != nil {
				return ServiceTokenList{}, err
			}
			if t.CreatedAt, err = store.ParseTime(created); err != nil {
				return ServiceTokenList{}, err
			}
			if t.LastUsedAt, err = parseNullTime(lastUsed); err != nil {
				return ServiceTokenList{}, err
			}
			if t.RevokedAt, err = parseNullTime(revok); err != nil {
				return ServiceTokenList{}, err
			}
			list.Tokens = append(list.Tokens, t)
		}
		return list, rows.Err()
	},
}

// RevokeServiceTokenInput is the input of revoke_service_token.
type RevokeServiceTokenInput struct {
	ID int64 `json:"id"`
}

// RevokeServiceToken revokes a service token at once.
var RevokeServiceToken = &operation.Op[RevokeServiceTokenInput, ServiceTokenInfo]{
	Name:        "revoke_service_token",
	Title:       "Revoke a service token",
	Description: "Revoke a service token. The program that used it loses access to the admin API at once.",
	Level:       operation.ServerAdmin,
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, in RevokeServiceTokenInput) (ServiceTokenInfo, error) {
		c.Log(fmt.Sprintf("service token %d", in.ID), "")
		t := ServiceTokenInfo{ID: in.ID}
		var created string
		var revoked sql.NullString
		err := c.Tx.QueryRowContext(ctx, "SELECT name, created_at, revoked_at FROM service_tokens WHERE id = ?", in.ID).
			Scan(&t.Name, &created, &revoked)
		if errors.Is(err, sql.ErrNoRows) {
			return ServiceTokenInfo{}, operation.NotFoundf("no service token %d", in.ID)
		}
		if err != nil {
			return ServiceTokenInfo{}, err
		}
		c.Log(fmt.Sprintf("service token %d", in.ID), t.Name)
		if revoked.Valid {
			return ServiceTokenInfo{}, operation.Invalidf("service token %d is already revoked", in.ID)
		}
		if t.CreatedAt, err = store.ParseTime(created); err != nil {
			return ServiceTokenInfo{}, err
		}
		now := c.Now
		if _, err := c.Tx.ExecContext(ctx, "UPDATE service_tokens SET revoked_at = ? WHERE id = ?", store.FormatTime(now), in.ID); err != nil {
			return ServiceTokenInfo{}, err
		}
		t.RevokedAt = &now
		return t, nil
	},
}

// ServiceOwner is the service token that a call presents.
type ServiceOwner struct {
	TokenID int64
	Name    string
}

// VerifyService checks a service token and returns it. It records when the
// token was last used, at most once a minute. It returns ErrInvalidToken
// when the token is unknown or revoked.
func VerifyService(ctx context.Context, db *store.DB, token string, now time.Time) (ServiceOwner, error) {
	if !strings.HasPrefix(token, ServiceTokenPrefix) {
		return ServiceOwner{}, ErrInvalidToken
	}
	var o ServiceOwner
	err := db.Read.QueryRowContext(ctx, "SELECT id, name FROM service_tokens WHERE hash = ? AND revoked_at IS NULL",
		credential.Hash(token)).Scan(&o.TokenID, &o.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return ServiceOwner{}, ErrInvalidToken
	}
	if err != nil {
		return ServiceOwner{}, err
	}
	_, err = db.Write.ExecContext(ctx, `
		UPDATE service_tokens SET last_used_at = ? WHERE id = ? AND (last_used_at IS NULL OR last_used_at < ?)`,
		store.FormatTime(now), o.TokenID, store.FormatTime(now.Add(-time.Minute)))
	return o, err
}

// Actor returns the actor of a call with the service token o, through the
// admin API, for the Cahut at domain.
func (o ServiceOwner) Actor(domain string) operation.Actor {
	return operation.Actor{Kind: operation.Service, Via: "admin-api", Domain: domain, Client: "service token " + o.Name}
}
