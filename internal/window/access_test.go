package window

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
)

func TestSetAccess(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatal(err)
	}
	w := out.(Window)
	if w.Access != AccessRoom {
		t.Errorf("a new Window opens to %q, want room", w.Access)
	}
	set := func(who operation.Actor, args string) (AccessSet, error) {
		t.Helper()
		out, err := f.invoke(t, who, "set_window_access", `{"window": "`+w.URL+`", `+args+`}`)
		if err != nil {
			return AccessSet{}, err
		}
		return out.(AccessSet), nil
	}

	// An AI opens a Window to its Cahut, or with a Key it draws.
	a, err := set(f.olivier, `"access": "internal"`)
	if err != nil || a.Access != AccessInternal || a.Key != "" {
		t.Errorf("internal = %+v, %v", a, err)
	}
	a, err = set(f.olivier, `"access": "key"`)
	if err != nil || a.Access != AccessKey || !regexp.MustCompile(`^[a-z2-9]{4}-[a-z2-9]{4}-[a-z2-9]{4}$`).MatchString(a.Key) {
		t.Fatalf("key = %+v, %v", a, err)
	}
	drawn := a.Key
	s, err := Lookup(ctx, f.db, "olivier", w.Name, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := OpenWithKey(ctx, f.db, s, "wrong key!", time.Now()); !errors.Is(err, ErrWrongKey) {
		t.Errorf("a wrong Key: err = %v, want ErrWrongKey", err)
	}
	token, expires, err := OpenWithKey(ctx, f.db, s, " "+drawn+" ", time.Now())
	if err != nil || !strings.HasPrefix(token, PassPrefix) || time.Until(expires) < PassTTL-time.Minute {
		t.Fatalf("OpenWithKey = %q, %v, %v", token, expires, err)
	}
	if pass, err := Pass(ctx, f.db, s.ID, token, time.Now()); err != nil || pass == 0 {
		t.Errorf("Pass = %d, %v", pass, err)
	}

	// Setting key again keeps the Key and its passes; a new Key ends them.
	if a, err := set(f.olivier, `"access": "key"`); err != nil || a.Key != "" {
		t.Errorf("key again = %+v, %v; want no new Key", a, err)
	}
	if pass, _ := Pass(ctx, f.db, s.ID, token, time.Now()); pass == 0 {
		t.Error("setting the same Key ended its passes")
	}
	a, err = set(f.olivier, `"access": "key", "new_key": true`)
	if err != nil || a.Key == "" || a.Key == drawn {
		t.Errorf("new_key = %+v, %v", a, err)
	}
	if pass, _ := Pass(ctx, f.db, s.ID, token, time.Now()); pass != 0 {
		t.Error("a new Key leaves the passes of the old one")
	}
	// The pass of a Key that changed after it was read is not given.
	if _, _, err := OpenWithKey(ctx, f.db, s, drawn, time.Now()); !errors.Is(err, ErrWrongKey) {
		t.Errorf("the old Key: err = %v, want ErrWrongKey", err)
	}
	if a, err := set(f.olivier, `"access": "key", "key": "correct horse battery"`); err != nil || a.Key != "correct horse battery" {
		t.Errorf("a chosen Key = %+v, %v", a, err)
	}
	var hash string
	if err := f.db.Read.QueryRowContext(ctx, "SELECT key_hash FROM windows WHERE id = ?", s.ID).Scan(&hash); err != nil || !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("key_hash = %q, %v", hash, err)
	}

	// The access may end, and then opens to the Room only.
	ends := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)
	a, err = set(f.olivier, `"access": "internal", "ends_at": "`+ends.Format(time.RFC3339)+`"`)
	if err != nil || a.AccessEndsAt == nil || !a.AccessEndsAt.Equal(ends) {
		t.Fatalf("internal until %v = %+v, %v", ends, a, err)
	}
	var keyHash *string
	if err := f.db.Read.QueryRowContext(ctx, "SELECT key_hash FROM windows WHERE id = ?", s.ID).Scan(&keyHash); err != nil || keyHash != nil {
		t.Errorf("the Key stays after the access changed: %v, %v", keyHash, err)
	}
	if s, err := Lookup(ctx, f.db, "olivier", w.Name, ends.Add(time.Second)); err != nil || s.Access != AccessRoom || s.EndsAt != nil {
		t.Errorf("after its end, the access is %q until %v, %v", s.Access, s.EndsAt, err)
	}
	out, _ = f.invoke(t, f.olivier, "get_window", `{"window": "`+w.URL+`"}`)
	if d := out.(Detail); d.Access != AccessInternal || d.AccessEndsAt == nil {
		t.Errorf("get_window = %+v", d.Window)
	}

	// What an AI cannot do, and what nobody can.
	if _, err := set(f.olivier, `"access": "public"`); !errors.Is(err, operation.ErrApprovalRequired) {
		t.Errorf("public by an AI: err = %v, want ErrApprovalRequired", err)
	}
	person := operation.Actor{Kind: operation.Person, Via: "api", UserID: f.olivier.UserID, Handle: "olivier", Session: 1}
	if a, err := set(person, `"access": "public"`); err != nil || a.Access != AccessPublic {
		t.Errorf("public by a person = %+v, %v", a, err)
	}
	for args, want := range map[string]error{
		`"access": "everyone"`:                                             operation.ErrInvalid,
		`"access": "room", "ends_at": "2099-01-01T00:00:00Z"`:              operation.ErrInvalid,
		`"access": "internal", "ends_at": "2001-01-01T00:00:00Z"`:          operation.ErrInvalid,
		`"access": "internal", "key": "correct horse battery"`:             operation.ErrInvalid,
		`"access": "key", "key": "short"`:                                  operation.ErrInvalid,
		`"access": "key", "key": "correct horse battery", "new_key": true`: operation.ErrInvalid,
	} {
		if _, err := set(person, args); !errors.Is(err, want) {
			t.Errorf("%s: err = %v, want %v", args, err, want)
		}
	}
	if _, err := set(f.camille, `"access": "internal"`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("another person: err = %v, want ErrNotFound", err)
	}

	// The Logbook says what changed, never the Key.
	entries, err := logbook.List(ctx, f.db.Read, 100)
	if err != nil {
		t.Fatal(err)
	}
	var keys int
	for _, e := range entries {
		if e.Operation != "set_window_access" {
			continue
		}
		if strings.Contains(e.Detail, drawn) || strings.Contains(e.Detail, "horse") || strings.Contains(e.Error, "horse") {
			t.Errorf("the Logbook holds a Key: %+v", e)
		}
		if e.Outcome == logbook.OK && strings.Contains(e.Detail, "new Key") {
			keys++
		}
	}
	if keys != 3 {
		t.Errorf("%d entries with a new Key, want 3", keys)
	}
}

func TestOpens(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatal(err)
	}
	name := out.(Window).Name
	tests := []struct {
		access       string
		userID, pass int64
		want         Opening
	}{
		{AccessRoom, 0, 0, Closed},
		{AccessRoom, f.camille.UserID, 0, Closed},
		{AccessRoom, f.olivier.UserID, 0, BySession},
		{AccessInternal, 0, 0, Closed},
		{AccessInternal, f.camille.UserID, 0, BySession},
		{AccessKey, 0, 7, ByPass},
		{AccessKey, f.camille.UserID, 0, Closed},
		{AccessKey, f.olivier.UserID, 7, BySession},
		{AccessPublic, 0, 0, ByAddress},
	}
	for _, tt := range tests {
		if _, err := f.db.Write.ExecContext(ctx, "UPDATE windows SET access = ?", tt.access); err != nil {
			t.Fatal(err)
		}
		w, err := Lookup(ctx, f.db, "olivier", name, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		if got, err := w.Opens(ctx, f.db, tt.userID, tt.pass); err != nil || got != tt.want {
			t.Errorf("%s for user %d with pass %d = %v, %v; want %v", tt.access, tt.userID, tt.pass, got, err, tt.want)
		}
	}
}
