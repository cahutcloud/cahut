package server

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/update"
)

// withUpdates opens a server whose updater keeps its state in a directory,
// as install.sh sets it up.
func withUpdates(t *testing.T) (*Server, string) {
	t.Helper()
	data, state := t.TempDir(), t.TempDir()
	s, err := Open(context.Background(), Config{DataDir: data, UpdateDir: state})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s, state
}

func writeStatus(t *testing.T, dir string, st update.Status) {
	t.Helper()
	data, err := json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, update.StatusFile), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestGetUpdates(t *testing.T) {
	ctx := context.Background()
	s := setup(t) // without an updater
	got, err := do[UpdatesInput, Updates](ctx, s, "get_updates", UpdatesInput{})
	if err != nil || got.Managed || got.Running == "" {
		t.Errorf("get_updates without an updater = %+v, %v", got, err)
	}

	s, state := withUpdates(t)
	at := time.Date(2026, 10, 23, 4, 0, 0, 0, time.UTC)
	writeStatus(t, state, update.Status{
		Policy: update.Patch,
		Check:  &update.Check{At: at, Latest: "0.5.0"},
		Last:   &update.Run{ID: "a", From: "0.4.0", To: "0.4.1", Result: update.Updated, Ended: at},
	})
	got, err = do[UpdatesInput, Updates](ctx, s, "get_updates", UpdatesInput{})
	if err != nil || !got.Managed || got.Policy != update.Patch || got.Check.Latest != "0.5.0" || got.Last.To != "0.4.1" {
		t.Errorf("get_updates = %+v, %v", got, err)
	}
}

// do runs the operation name of the server's registry, as the command line.
// The operations on updates are made by Open, as closures on the server.
// A visitor may be given instead.
func do[In, Out any](ctx context.Context, s *Server, name string, in In, as ...operation.Actor) (Out, error) {
	var out Out
	data, err := json.Marshal(in)
	if err != nil {
		return out, err
	}
	actor := admin
	if len(as) > 0 {
		actor = as[0]
	}
	res, err := s.Registry.Invoke(ctx, name, actor, data)
	if err != nil {
		return out, err
	}
	b, err := json.Marshal(res)
	if err != nil {
		return out, err
	}
	return out, json.Unmarshal(b, &out)
}

func TestRequestUpdate(t *testing.T) {
	ctx := context.Background()
	s := setup(t)
	if _, err := do[RequestUpdateInput, RequestedUpdate](ctx, s, "request_update", RequestUpdateInput{}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("request_update on a server that does not update itself: err = %v", err)
	}

	s, state := withUpdates(t)
	writeStatus(t, state, update.Status{Policy: update.Patch})
	// A test binary runs version dev, which any release is newer than: the
	// updater refuses to go back, and its tests check it.
	for _, v := range []string{"next", "0.4", "latest"} {
		if _, err := do[RequestUpdateInput, RequestedUpdate](ctx, s, "request_update", RequestUpdateInput{Version: v}); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("request_update %q: err = %v, want ErrInvalid", v, err)
		}
	}
	got, err := do[RequestUpdateInput, RequestedUpdate](ctx, s, "request_update", RequestUpdateInput{Version: "v99.0.0"})
	if err != nil || got.Requested != "99.0.0" {
		t.Fatalf("request_update 99.0.0 = %+v, %v", got, err)
	}
	if what, asked, err := update.ReadRequest(s.cfg.DataDir); what != "99.0.0" || !asked || err != nil {
		t.Errorf("the request holds %q, %v, %v", what, asked, err)
	}
	u, err := do[UpdatesInput, Updates](ctx, s, "get_updates", UpdatesInput{})
	if err != nil || u.Requested != "99.0.0" {
		t.Errorf("get_updates = %+v, %v; want the request shown", u, err)
	}
	if _, err := do[RequestUpdateInput, RequestedUpdate](ctx, s, "request_update", RequestUpdateInput{}); err != nil {
		t.Fatal(err)
	}
	if u, _ := do[UpdatesInput, Updates](ctx, s, "get_updates", UpdatesInput{}); u.Requested != "check" {
		t.Errorf("after a check is asked for, requested = %q", u.Requested)
	}

	entries, err := logbook.List(ctx, s.DB.Read, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 || entries[0].Operation != "request_update" || entries[0].Target != "check" {
		t.Errorf("latest entry = %+v", entries)
	}
}

func TestLogUpdates(t *testing.T) {
	ctx := context.Background()
	s, state := withUpdates(t)
	if err := s.LogUpdates(ctx); err != nil {
		t.Fatalf("LogUpdates with no update yet: %v", err)
	}
	at := time.Date(2026, 10, 23, 4, 0, 0, 0, time.UTC)
	writeStatus(t, state, update.Status{Last: &update.Run{
		ID: "one", From: "0.4.0", To: "0.4.1", By: update.ByTimer, Result: update.RolledBack,
		Reason: "the Cahut of a.example did not open", Ended: at,
	}})
	for range 2 { // the second time finds it recorded
		if err := s.LogUpdates(ctx); err != nil {
			t.Fatal(err)
		}
	}
	writeStatus(t, state, update.Status{Last: &update.Run{
		ID: "two", From: "0.4.0", To: "0.4.2", By: update.ByConsole, Result: update.Updated, Ended: at.Add(time.Hour),
	}})
	if err := s.LogUpdates(ctx); err != nil {
		t.Fatal(err)
	}

	// Two updates within a minute, which the history holds both of.
	writeStatus(t, state, update.Status{History: []update.Run{
		{ID: "four", From: "0.4.3", To: "0.4.4", Result: update.Updated, Ended: at.Add(3 * time.Hour)},
		{ID: "three", From: "0.4.2", To: "0.4.3", Result: update.Updated, Ended: at.Add(2 * time.Hour)},
		{ID: "two", From: "0.4.0", To: "0.4.2", By: update.ByConsole, Result: update.Updated, Ended: at.Add(time.Hour)},
		{ID: "running", From: "0.4.4", To: "0.4.5"}, // not ended yet
	}})
	if err := s.LogUpdates(ctx); err != nil {
		t.Fatal(err)
	}

	entries, err := logbook.List(ctx, s.DB.Read, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 || entries[0].Target != "0.4.3 to 0.4.4" || entries[1].Target != "0.4.2 to 0.4.3" {
		t.Fatalf("%d entries, want 4, the newest first: %+v", len(entries), entries)
	}
	entries = entries[2:]
	updated, undone := entries[0], entries[1]
	if updated.Operation != "update" || updated.Target != "0.4.0 to 0.4.2" || updated.Outcome != logbook.OK || updated.Via != "api" {
		t.Errorf("the update = %+v", updated)
	}
	if undone.Outcome != logbook.Failed || undone.Detail != "rolled back" || undone.Via != "timer" ||
		!strings.Contains(undone.Error, "did not open") || !undone.At.Equal(at) {
		t.Errorf("the update undone = %+v", undone)
	}
}
