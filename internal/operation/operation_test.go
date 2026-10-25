package operation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/store"
)

// The people of the test registry: 1 is a member, 2 an administrator of the
// Cahut, and 3 is disabled.
var (
	admin    = Actor{Kind: Admin, Via: "cli"}
	ai       = Actor{Kind: AI, Via: "mcp", UserID: 1, Handle: "olivier", TokenID: 7, Client: "Claude Code"}
	member   = Actor{Kind: Person, Via: "api", UserID: 1, Handle: "olivier", Session: 4}
	adminAI  = Actor{Kind: AI, Via: "mcp", UserID: 2, Handle: "camille", TokenID: 8, Client: "Claude Code"}
	disabled = Actor{Kind: Person, Via: "api", UserID: 3, Handle: "bruno", Session: 5}
	visitor  = Actor{Kind: Visitor, Via: "api"}
	service  = Actor{Kind: Service, Via: "admin-api", Client: "service token account"}
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
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "cahut.db"), store.CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := NewRegistry(db)
	r.Now = func() time.Time { return time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC) }
	r.CheckPerson = func(_ context.Context, _ *sql.Tx, userID int64) (bool, error) {
		if userID == 3 {
			return false, Forbiddenf("this person is disabled")
		}
		return userID == 2, nil
	}
	return r, db
}

func createRoom(fail error) *Op[roomIn, roomOut] {
	return &Op[roomIn, roomOut]{
		Name: "create_room",
		Run: func(ctx context.Context, c *Call, in roomIn) (roomOut, error) {
			res, err := c.Tx.ExecContext(ctx, "INSERT INTO rooms (name, kind, title, created_at) VALUES (?, 'team', ?, ?)",
				in.Name, in.Name, store.FormatTime(c.Now))
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

// rules copies the rules of an operation, not its name or its Run.
func rules(op *Op[roomIn, roomOut], from Op[roomIn, roomOut]) {
	op.Level, op.PeopleOnly, op.Visitors, op.KnockFirst = from.Level, from.PeopleOnly, from.Visitors, from.KnockFirst
	op.Services = from.Services
}

func TestRefusals(t *testing.T) {
	tests := []struct {
		name  string
		rules Op[roomIn, roomOut]
		actor Actor
		want  error
	}{
		{name: "for people, by an AI", rules: Op[roomIn, roomOut]{PeopleOnly: true}, actor: ai, want: ErrForbidden},
		{name: "knock first, by an AI", rules: Op[roomIn, roomOut]{KnockFirst: true}, actor: adminAI, want: ErrApprovalRequired},
		{name: "for Cahut administrators, by a member", rules: Op[roomIn, roomOut]{Level: CahutAdmin}, actor: member, want: ErrForbidden},
		{name: "for Cahut administrators, by the AI of a member", rules: Op[roomIn, roomOut]{Level: CahutAdmin}, actor: ai, want: ErrForbidden},
		{name: "for server administrators, by a person", rules: Op[roomIn, roomOut]{Level: ServerAdmin}, actor: member, want: ErrForbidden},
		{name: "by a disabled person", actor: disabled, want: ErrForbidden},
		{name: "by a visitor", actor: visitor, want: ErrForbidden},
		{name: "not open to services, by a service", rules: Op[roomIn, roomOut]{Level: CahutAdmin}, actor: service, want: ErrForbidden},
		{name: "for server administrators, by a service", rules: Op[roomIn, roomOut]{Level: ServerAdmin, Services: true}, actor: service, want: ErrForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, db := testRegistry(t)
			op := createRoom(nil)
			rules(op, tt.rules)
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

func TestServiceWithoutItsToken(t *testing.T) {
	r, db := testRegistry(t)
	op := createRoom(nil)
	op.Services = true
	Register(r, op)
	if _, err := Do(context.Background(), r, op, Actor{Kind: Service, Via: "admin-api"}, roomIn{Name: "Team"}); err == nil || Public(err) {
		t.Errorf("a service without its token: err = %v, want an internal error", err)
	}
	if rooms(t, db) != 0 {
		t.Error("the operation ran")
	}
}

func TestAdmission(t *testing.T) {
	tests := []struct {
		name      string
		rules     Op[roomIn, roomOut]
		actor     Actor
		wantAdmin bool
	}{
		{name: "for any person, by a member", actor: member},
		{name: "for any person, by the command line", actor: admin, wantAdmin: true},
		{name: "for Cahut administrators, by the AI of an administrator", rules: Op[roomIn, roomOut]{Level: CahutAdmin}, actor: adminAI, wantAdmin: true},
		{name: "for server administrators, by the command line", rules: Op[roomIn, roomOut]{Level: ServerAdmin}, actor: admin, wantAdmin: true},
		{name: "knock first, by a person", rules: Op[roomIn, roomOut]{KnockFirst: true}, actor: member},
		{name: "for visitors, by a visitor", rules: Op[roomIn, roomOut]{Visitors: true, PeopleOnly: true}, actor: visitor},
		{name: "open to services, knock first, by a service", rules: Op[roomIn, roomOut]{
			Level: CahutAdmin, KnockFirst: true, PeopleOnly: true, Services: true,
		}, actor: service, wantAdmin: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, db := testRegistry(t)
			op := createRoom(nil)
			rules(op, tt.rules)
			run := op.Run
			var gotAdmin bool
			op.Run = func(ctx context.Context, c *Call, in roomIn) (roomOut, error) {
				gotAdmin = c.Admin
				c.Identify(Person, 9, "someone")
				return run(ctx, c, in)
			}
			Register(r, op)

			if _, err := Do(context.Background(), r, op, tt.actor, roomIn{Name: "Team"}); err != nil {
				t.Fatalf("Do: %v", err)
			}
			if gotAdmin != tt.wantAdmin {
				t.Errorf("c.Admin = %v, want %v", gotAdmin, tt.wantAdmin)
			}
			got := entries(t, db)
			if len(got) != 1 {
				t.Fatalf("entries = %+v", got)
			}
			// A visitor is logged as the person Identify names; Identify
			// changes nobody else.
			// A service is logged as an administrator, under its token.
			wantActor, wantHandle := string(tt.actor.Kind), tt.actor.Handle
			switch tt.actor.Kind {
			case Visitor:
				wantActor, wantHandle = "person", "someone"
			case Service:
				wantActor = "admin"
			}
			if got[0].Actor != wantActor || got[0].UserHandle != wantHandle {
				t.Errorf("entry actor %q %q, want %q %q", got[0].Actor, got[0].UserHandle, wantActor, wantHandle)
			}
			if got[0].Via != tt.actor.Via || got[0].Client != tt.actor.Client {
				t.Errorf("entry via %q, client %q", got[0].Via, got[0].Client)
			}
		})
	}
}

func TestForAIs(t *testing.T) {
	tests := []struct {
		info          Info
		member, admin bool
	}{
		{info: Info{}, member: true, admin: true},
		{info: Info{KnockFirst: true}, member: true, admin: true},
		{info: Info{Level: CahutAdmin}, admin: true},
		{info: Info{Level: ServerAdmin}},
		{info: Info{PeopleOnly: true}},
		{info: Info{PeopleOnly: true, Level: CahutAdmin}},
	}
	for _, tt := range tests {
		if got := tt.info.ForAIs(false, false); got != tt.member {
			t.Errorf("%+v ForAIs(false, false) = %v", tt.info, got)
		}
		if got := tt.info.ForAIs(true, false); got != tt.admin {
			t.Errorf("%+v ForAIs(true, false) = %v", tt.info, got)
		}
		// An AI that may only read sees the operations that read.
		if got := tt.info.ForAIs(true, true); got {
			t.Errorf("%+v ForAIs(true, true) = %v", tt.info, got)
		}
		read := tt.info
		read.ReadOnly = true
		if got := read.ForAIs(true, true); got != tt.admin {
			t.Errorf("%+v ForAIs(true, true) = %v", read, got)
		}
		asks := tt.info
		asks.ReadOnlyAIs = true
		if got := asks.ForAIs(true, true); got != tt.admin {
			t.Errorf("%+v ForAIs(true, true) = %v", asks, got)
		}
	}
}

func TestForAIsIn(t *testing.T) {
	tests := []struct {
		info          Info
		member, admin bool // in an individual Cahut
	}{
		{info: Info{}, member: true, admin: true},
		{info: Info{InIndividual: CahutAdmin}, admin: true},
		{info: Info{InIndividual: Nobody}},
		{info: Info{Level: CahutAdmin, InIndividual: AnyPerson}, admin: true},
		{info: Info{Level: ServerAdmin, InIndividual: CahutAdmin}},
	}
	for _, tt := range tests {
		if got := tt.info.ForAIsIn(true, false, false); got != tt.member {
			t.Errorf("%+v ForAIsIn(true, false, false) = %v", tt.info, got)
		}
		if got := tt.info.ForAIsIn(true, true, false); got != tt.admin {
			t.Errorf("%+v ForAIsIn(true, true, false) = %v", tt.info, got)
		}
		// In a collaborative Cahut, InIndividual changes nothing.
		if got, want := tt.info.ForAIsIn(false, false, false), tt.info.ForAIs(false, false); got != want {
			t.Errorf("%+v ForAIsIn(false, false, false) = %v, want %v", tt.info, got, want)
		}
	}
}

// TestInIndividual checks that the registry applies the level of an
// operation in an individual Cahut, which it reads in the transaction of
// the call.
func TestInIndividual(t *testing.T) {
	r, _ := testRegistry(t)
	individual := false
	r.Individual = func(context.Context, *sql.Tx) (bool, error) { return individual, nil }
	list := &Op[roomIn, roomOut]{Name: "list_people", ReadOnly: true, InIndividual: CahutAdmin,
		Run: func(context.Context, *Call, roomIn) (roomOut, error) { return roomOut{}, nil }}
	add := &Op[roomIn, roomOut]{Name: "add_member", InIndividual: Nobody,
		Run: func(context.Context, *Call, roomIn) (roomOut, error) { return roomOut{}, nil }}
	Register(r, list)
	Register(r, add)
	ctx := context.Background()
	member := Actor{Kind: Person, Via: "api", UserID: 1}
	cahutAdmin := Actor{Kind: Person, Via: "api", UserID: 2} // testRegistry makes person 2 an administrator
	for _, who := range []Actor{member, cahutAdmin} {
		if _, err := Do(ctx, r, list, who, roomIn{}); err != nil {
			t.Errorf("list_people by %d in a collaborative Cahut: %v", who.UserID, err)
		}
		if _, err := Do(ctx, r, add, who, roomIn{}); err != nil {
			t.Errorf("add_member by %d in a collaborative Cahut: %v", who.UserID, err)
		}
	}
	individual = true
	if _, err := Do(ctx, r, list, member, roomIn{}); !errors.Is(err, ErrForbidden) || !strings.Contains(err.Error(), "which is individual") {
		t.Errorf("list_people by a member in an individual Cahut: %v", err)
	}
	if _, err := Do(ctx, r, list, cahutAdmin, roomIn{}); err != nil {
		t.Errorf("list_people by an administrator in an individual Cahut: %v", err)
	}
	if _, err := Do(ctx, r, add, cahutAdmin, roomIn{}); !errors.Is(err, ErrForbidden) || !strings.Contains(err.Error(), "each Room keeps one member") {
		t.Errorf("add_member by an administrator in an individual Cahut: %v", err)
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
