package operation

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/store"
)

var (
	admin = Actor{Kind: Admin, Via: "cli"}
	ai    = Actor{Kind: AI, Via: "mcp", UserID: 1, Handle: "olivier", TokenID: 7, Client: "Claude Code"}
)

type roomIn struct {
	Name string `json:"name"`
}

type roomOut struct {
	ID int64 `json:"id"`
}

// testRegistry returns a registry with operations that create and count
// rooms, a table that needs no other row.
func testRegistry(t *testing.T) (*Registry, *store.DB) {
	t.Helper()
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "cahut.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := NewRegistry(db)
	r.Now = func() time.Time { return time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC) }
	return r, db
}

func createRoom(fail error) *Op[roomIn, roomOut] {
	return &Op[roomIn, roomOut]{
		Name: "create_room",
		Run: func(ctx context.Context, c *Call, in roomIn) (roomOut, error) {
			res, err := c.Tx.ExecContext(ctx, "INSERT INTO rooms (kind, name, created_at) VALUES ('team', ?, ?)",
				in.Name, store.FormatTime(c.Now))
			if err != nil {
				return roomOut{}, err
			}
			c.Log(in.Name, "a detail")
			if fail != nil {
				return roomOut{}, fail
			}
			id, err := res.LastInsertId()
			return roomOut{ID: id}, err
		},
	}
}

var countRooms = &Op[struct{}, int]{
	Name:     "count_rooms",
	ReadOnly: true,
	Run: func(ctx context.Context, c *Call, _ struct{}) (int, error) {
		var n int
		err := c.Tx.QueryRowContext(ctx, "SELECT count(*) FROM rooms").Scan(&n)
		return n, err
	},
}

func rooms(t *testing.T, db *store.DB) int {
	t.Helper()
	var n int
	if err := db.Read.QueryRowContext(context.Background(), "SELECT count(*) FROM rooms").Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func entries(t *testing.T, db *store.DB) []logbook.Entry {
	t.Helper()
	e, err := logbook.List(context.Background(), db.Read, 100)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestChangeIsLoggedWithIt(t *testing.T) {
	r, db := testRegistry(t)
	op := createRoom(nil)
	Register(r, op)
	committed := false
	run := op.Run
	op.Run = func(ctx context.Context, c *Call, in roomIn) (roomOut, error) {
		c.AfterCommit(func() { committed = true })
		return run(ctx, c, in)
	}

	out, err := Do(context.Background(), r, op, ai, roomIn{Name: "Team"})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if out.ID == 0 || rooms(t, db) != 1 {
		t.Errorf("room not created: out = %+v, rooms = %d", out, rooms(t, db))
	}
	if !committed {
		t.Error("AfterCommit did not run")
	}
	got := entries(t, db)
	if len(got) != 1 {
		t.Fatalf("got %d entries, want 1", len(got))
	}
	e := got[0]
	if e.Operation != "create_room" || e.Actor != "ai" || e.UserHandle != "olivier" || e.TokenID != 7 ||
		e.Client != "Claude Code" || e.Via != "mcp" || e.Target != "Team" || e.Detail != "a detail" ||
		e.Outcome != logbook.OK || e.Error != "" || !e.At.Equal(r.Now()) {
		t.Errorf("entry = %+v", e)
	}
}

func TestFailedChangeIsRolledBackAndLogged(t *testing.T) {
	r, db := testRegistry(t)
	op := createRoom(Invalidf("bad name"))
	Register(r, op)
	cleaned, committed := false, false
	run := op.Run
	op.Run = func(ctx context.Context, c *Call, in roomIn) (roomOut, error) {
		c.OnFailure(func() { cleaned = true })
		c.AfterCommit(func() { committed = true })
		return run(ctx, c, in)
	}

	_, err := Do(context.Background(), r, op, admin, roomIn{Name: "Team"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
	if rooms(t, db) != 0 {
		t.Error("the change was not rolled back")
	}
	if !cleaned || committed {
		t.Errorf("OnFailure ran = %v, AfterCommit ran = %v", cleaned, committed)
	}
	got := entries(t, db)
	if len(got) != 1 || got[0].Outcome != logbook.Failed || got[0].Error != "invalid input: bad name" || got[0].Target != "Team" {
		t.Errorf("entries = %+v", got)
	}
}

func TestReadsAreLoggedForAIsOnly(t *testing.T) {
	r, db := testRegistry(t)
	Register(r, countRooms)

	if _, err := Do(context.Background(), r, countRooms, admin, struct{}{}); err != nil {
		t.Fatal(err)
	}
	if n := len(entries(t, db)); n != 0 {
		t.Errorf("admin read: %d entries, want 0", n)
	}
	if _, err := Do(context.Background(), r, countRooms, ai, struct{}{}); err != nil {
		t.Fatal(err)
	}
	got := entries(t, db)
	if len(got) != 1 || got[0].Operation != "count_rooms" || got[0].Outcome != logbook.OK {
		t.Errorf("AI read: entries = %+v", got)
	}
}

func TestRefusals(t *testing.T) {
	tests := []struct {
		name  string
		op    *Op[roomIn, roomOut]
		actor Actor
		want  error
	}{
		{name: "admin only, by an AI", op: &Op[roomIn, roomOut]{AdminOnly: true}, actor: ai, want: ErrForbidden},
		{name: "irreversible, by an AI", op: &Op[roomIn, roomOut]{Irreversible: true}, actor: ai, want: ErrApprovalRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, db := testRegistry(t)
			op := createRoom(nil)
			op.AdminOnly, op.Irreversible = tt.op.AdminOnly, tt.op.Irreversible
			Register(r, op)

			_, err := Do(context.Background(), r, op, tt.actor, roomIn{Name: "Team"})
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if !Public(err) {
				t.Error("the refusal is not public")
			}
			if rooms(t, db) != 0 {
				t.Error("the operation ran")
			}
			got := entries(t, db)
			if len(got) != 1 || got[0].Outcome != logbook.Refused {
				t.Errorf("entries = %+v", got)
			}

			if _, err := Do(context.Background(), r, op, admin, roomIn{Name: "Team"}); err != nil {
				t.Errorf("the administrator was refused: %v", err)
			}
		})
	}
}

func TestInvoke(t *testing.T) {
	r, db := testRegistry(t)
	Register(r, createRoom(nil))
	ctx := context.Background()

	out, err := r.Invoke(ctx, "create_room", ai, json.RawMessage(`{"name": "Team"}`))
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if out.(roomOut).ID == 0 {
		t.Errorf("out = %+v", out)
	}

	_, err = r.Invoke(ctx, "create_room", ai, json.RawMessage(`{"name": "Team", "size": 3}`))
	if !errors.Is(err, ErrInvalid) {
		t.Errorf("unknown field: err = %v, want ErrInvalid", err)
	}
	if got := entries(t, db); len(got) != 2 || got[0].Outcome != logbook.Failed {
		t.Errorf("the invalid call was not logged: %+v", got)
	}

	if _, err := r.Invoke(ctx, "delete_everything", ai, nil); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown operation: err = %v, want ErrNotFound", err)
	}
}

func TestRegister(t *testing.T) {
	r, _ := testRegistry(t)
	Register(r, countRooms)
	Register(r, createRoom(nil))

	infos := r.Ops()
	if len(infos) != 2 || infos[0].Name != "count_rooms" || infos[1].Name != "create_room" {
		t.Errorf("Ops() = %+v", infos)
	}
	if !infos[0].ReadOnly || infos[0].Output.Kind().String() != "int" {
		t.Errorf("count_rooms info = %+v", infos[0])
	}

	for _, name := range []string{"count_rooms", "Bad-Name"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Register(%q) did not panic", name)
				}
			}()
			Register(r, &Op[struct{}, int]{Name: name})
		}()
	}
}
