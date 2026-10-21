package account

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

func setup(t *testing.T) (*operation.Registry, *store.DB) {
	t.Helper()
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "cahut.db"), store.CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := operation.NewRegistry(db)
	Register(r)
	return r, db
}

func TestCreateUser(t *testing.T) {
	r, db := setup(t)
	ctx := context.Background()

	u, err := operation.Do(ctx, r, CreateUser, admin, CreateUserInput{Handle: "olivier", Name: " Olivier Durand ", Admin: true})
	if err != nil {
		t.Fatalf("create_user: %v", err)
	}
	if u.ID == 0 || u.Name != "Olivier Durand" || u.Role != "admin" || u.PersonalRoom != "olivier" {
		t.Errorf("user = %+v", u)
	}
	var role, room, title string
	err = db.Read.QueryRowContext(ctx, `
		SELECT m.role, r.name, r.title FROM room_members m JOIN rooms r ON r.id = m.room_id
		WHERE m.user_id = ? AND r.kind = 'personal'`, u.ID).Scan(&role, &room, &title)
	if err != nil || role != "owner" || room != "olivier" || title != "Olivier Durand" {
		t.Errorf("personal Room: %q %q, role %q, err %v", room, title, role, err)
	}

	if _, err := db.Write.ExecContext(ctx, "INSERT INTO rooms (name, kind, title, created_at) VALUES ('marketing', 'team', 'Marketing', 'now')"); err != nil {
		t.Fatal(err)
	}
	for _, in := range []CreateUserInput{
		{Handle: "olivier", Name: "Someone else"},
		{Handle: "marketing", Name: "Taken by a Room"},
		{Handle: "mcp", Name: "Reserved"},
		{Handle: "Olivier", Name: "Capital"},
		{Handle: "x", Name: "Too short"},
		{Handle: "camille-", Name: "Ends with a hyphen"},
		{Handle: "camille", Name: "  "},
	} {
		if _, err := operation.Do(ctx, r, CreateUser, admin, in); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("create_user(%+v): err = %v, want ErrInvalid", in, err)
		}
	}
}

func TestTokens(t *testing.T) {
	r, db := setup(t)
	ctx := context.Background()
	if _, err := operation.Do(ctx, r, CreateUser, admin, CreateUserInput{Handle: "olivier", Name: "Olivier"}); err != nil {
		t.Fatal(err)
	}

	nt, err := operation.Do(ctx, r, CreateToken, admin, CreateTokenInput{User: "olivier", Name: "Claude Code"})
	if err != nil {
		t.Fatalf("create_token: %v", err)
	}
	if !strings.HasPrefix(nt.Token, TokenPrefix) || len(nt.Token) != len(TokenPrefix)+52 {
		t.Errorf("token = %q", nt.Token)
	}
	if _, err := operation.Do(ctx, r, CreateToken, admin, CreateTokenInput{User: "nobody", Name: "x"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("token for an unknown user: err = %v, want ErrNotFound", err)
	}

	entries, err := logbook.List(ctx, db.Read, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Target+e.Detail+e.Error, TokenPrefix) {
			t.Errorf("the Logbook holds a token: %+v", e)
		}
	}

	now := time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)
	o, err := Verify(ctx, db, nt.Token, now)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if o.TokenID != nt.ID || o.Handle != "olivier" || o.TokenName != "Claude Code" {
		t.Errorf("owner = %+v", o)
	}
	for _, bad := range []string{"", "cahut_pat_nope", nt.Token + "x", strings.TrimPrefix(nt.Token, TokenPrefix)} {
		if _, err := Verify(ctx, db, bad, now); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("Verify(%q): err = %v, want ErrInvalidToken", bad, err)
		}
	}

	list, err := operation.Do(ctx, r, ListTokens, admin, ListTokensInput{User: "olivier"})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tokens) != 1 || list.Tokens[0].LastUsedAt == nil || !list.Tokens[0].LastUsedAt.Equal(now) {
		t.Errorf("list_tokens = %+v", list)
	}

	if _, err := operation.Do(ctx, r, RevokeToken, admin, RevokeTokenInput{ID: nt.ID}); err != nil {
		t.Fatalf("revoke_token: %v", err)
	}
	if _, err := Verify(ctx, db, nt.Token, now); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Verify after revoke: err = %v, want ErrInvalidToken", err)
	}
	if _, err := operation.Do(ctx, r, RevokeToken, admin, RevokeTokenInput{ID: nt.ID}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("second revoke: err = %v, want ErrInvalid", err)
	}
}
