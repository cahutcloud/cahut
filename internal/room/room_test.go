package room

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

var cli = operation.Actor{Kind: operation.Admin, Via: "cli"}

// fixture is a Cahut with olivier, an administrator, and camille and
// bruno, members. Each acts in the console.
type fixture struct {
	r                       *operation.Registry
	db                      *store.DB
	olivier, camille, bruno operation.Actor
}

func setup(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "cahut.db"), store.CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	f := &fixture{r: operation.NewRegistry(db), db: db}
	account.Register(f.r)
	Register(f.r)
	for _, p := range []struct {
		handle string
		to     *operation.Actor
	}{{"olivier", &f.olivier}, {"camille", &f.camille}, {"bruno", &f.bruno}} {
		u, err := operation.Do(ctx, f.r, account.CreateUser, cli, account.CreateUserInput{Handle: p.handle, Name: p.handle, Admin: p.handle == "olivier"})
		if err != nil {
			t.Fatal(err)
		}
		*p.to = operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: p.handle, Session: 1}
	}
	return f
}

// ai is the AI of the person p.
func ai(p operation.Actor) operation.Actor {
	return operation.Actor{Kind: operation.AI, Via: "mcp", UserID: p.UserID, Handle: p.Handle, TokenID: 1, Client: "Claude Code"}
}

func (f *fixture) people(t *testing.T, who operation.Actor, room string) []string {
	t.Helper()
	d, err := operation.Do(context.Background(), f.r, GetRoom, who, GetInput{Room: room})
	if err != nil {
		t.Fatalf("get_room %s: %v", room, err)
	}
	var got []string
	for _, p := range d.People {
		got = append(got, p.Handle+" "+p.Role)
	}
	return got
}

func TestCreateRoom(t *testing.T) {
	f := setup(t)
	ctx := context.Background()

	r, err := operation.Do(ctx, f.r, CreateRoom, f.camille, CreateRoomInput{Title: " Équipe Marketing "})
	if err != nil {
		t.Fatalf("create_room: %v", err)
	}
	if r.Name != "equipe-marketing" || r.Title != "Équipe Marketing" || r.Kind != Team || r.Role != Owner || r.Members != 1 {
		t.Errorf("room = %+v", r)
	}
	if r, err := operation.Do(ctx, f.r, CreateRoom, ai(f.bruno), CreateRoomInput{Title: "Sales", Name: "sales"}); err != nil || r.Role != Owner {
		t.Errorf("create_room by an AI = %+v, %v", r, err)
	}
	if r, err := operation.Do(ctx, f.r, CreateRoom, cli, CreateRoomInput{Title: "Board", Owner: "olivier"}); err != nil || r.Name != "board" || r.Role != "" {
		t.Errorf("create_room by the command line = %+v, %v", r, err)
	}

	for in, want := range map[CreateRoomInput]error{
		{Title: "Marketing", Name: "equipe-marketing"}: operation.ErrInvalid, // taken by a Room
		{Title: "Camille"}:               operation.ErrInvalid,   // taken by a person
		{Title: "Help"}:                  operation.ErrInvalid,   // reserved
		{Title: "x", Name: "Bad_Name"}:   operation.ErrInvalid,   // not a name
		{Title: "!!!"}:                   operation.ErrInvalid,   // no name in the title
		{Title: " "}:                     operation.ErrInvalid,   // no title
		{Title: "Ops", Owner: "olivier"}: operation.ErrForbidden, // a member names another owner
		{Title: "Ops", Owner: "nobody"}:  operation.ErrForbidden,
	} {
		if _, err := operation.Do(ctx, f.r, CreateRoom, f.camille, in); !errors.Is(err, want) {
			t.Errorf("create_room(%+v): err = %v, want %v", in, err, want)
		}
	}
	if _, err := operation.Do(ctx, f.r, CreateRoom, cli, CreateRoomInput{Title: "Ops"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("create_room by the command line, without an owner: err = %v, want ErrInvalid", err)
	}
	if _, err := operation.Do(ctx, f.r, CreateRoom, f.olivier, CreateRoomInput{Title: "Ops", Owner: "nobody"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("create_room for nobody: err = %v, want ErrNotFound", err)
	}
}

func TestListAndGetRooms(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	if _, err := operation.Do(ctx, f.r, CreateRoom, f.camille, CreateRoomInput{Title: "Marketing"}); err != nil {
		t.Fatal(err)
	}
	names := func(l List) []string {
		var got []string
		for _, r := range l.Rooms {
			got = append(got, r.Name+" "+r.Role)
		}
		return got
	}
	l, err := operation.Do(ctx, f.r, ListRooms, f.camille, ListRoomsInput{})
	if err != nil || !slices.Equal(names(l), []string{"camille owner", "marketing owner"}) {
		t.Errorf("list_rooms by camille = %q, %v", names(l), err)
	}
	if _, err := operation.Do(ctx, f.r, ListRooms, f.camille, ListRoomsInput{All: true}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("list_rooms all by a member: err = %v, want ErrForbidden", err)
	}
	l, err = operation.Do(ctx, f.r, ListRooms, f.olivier, ListRoomsInput{All: true})
	if want := []string{"olivier owner", "bruno ", "camille ", "marketing "}; err != nil || !slices.Equal(names(l), want) {
		t.Errorf("list_rooms all by an administrator = %q, %v; want %q", names(l), err, want)
	}

	if got := f.people(t, f.camille, "marketing"); !slices.Equal(got, []string{"camille owner"}) {
		t.Errorf("people of marketing = %q", got)
	}
	// An administrator sees every Room; anyone else, theirs only.
	f.people(t, f.olivier, "marketing")
	if _, err := operation.Do(ctx, f.r, GetRoom, f.bruno, GetInput{Room: "marketing"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_room by someone outside: err = %v, want ErrNotFound", err)
	}
	if _, err := operation.Do(ctx, f.r, GetRoom, f.bruno, GetInput{Room: "nope"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_room of no Room: err = %v, want ErrNotFound", err)
	}

	if _, err := operation.Do(ctx, f.r, UpdateRoom, f.bruno, UpdateRoomInput{Room: "marketing", Title: "Ours"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("update_room by someone outside: err = %v, want ErrNotFound", err)
	}
	d, err := operation.Do(ctx, f.r, UpdateRoom, f.olivier, UpdateRoomInput{Room: "marketing", Title: " Marketing and sales "})
	if err != nil || d.Title != "Marketing and sales" || d.Name != "marketing" {
		t.Errorf("update_room by an administrator = %+v, %v", d, err)
	}
}

func TestMembers(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	if _, err := operation.Do(ctx, f.r, CreateRoom, f.camille, CreateRoomInput{Title: "Marketing"}); err != nil {
		t.Fatal(err)
	}
	add := MemberInput{Room: "marketing", User: "bruno"}
	if _, err := operation.Do(ctx, f.r, AddMember, ai(f.camille), add); !errors.Is(err, operation.ErrApprovalRequired) {
		t.Errorf("add_room_member by an AI: err = %v, want ErrApprovalRequired", err)
	}
	if _, err := operation.Do(ctx, f.r, AddMember, f.camille, add); err != nil {
		t.Fatalf("add_room_member: %v", err)
	}
	for in, want := range map[MemberInput]error{
		add: operation.ErrInvalid, // already in
		{Room: "marketing", User: "olivier", Role: "boss"}: operation.ErrInvalid,
		{Room: "marketing", User: "nobody"}:                operation.ErrNotFound,
		{Room: "camille", User: "bruno"}:                   operation.ErrInvalid, // a personal Room
	} {
		if _, err := operation.Do(ctx, f.r, AddMember, f.camille, in); !errors.Is(err, want) {
			t.Errorf("add_room_member(%+v): err = %v, want %v", in, err, want)
		}
	}
	// A member publishes, but does not manage the Room.
	if _, err := operation.Do(ctx, f.r, AddMember, f.bruno, MemberInput{Room: "marketing", User: "olivier"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("add_room_member by a member: err = %v, want ErrForbidden", err)
	}
	if _, err := Require(ctx, call(t, f, f.bruno), "marketing", Member); err != nil {
		t.Errorf("Require member: %v", err)
	}
	// An administrator manages the Room, but does not publish in it.
	if _, err := Require(ctx, call(t, f, f.olivier), "marketing", Member); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("Require member, for an administrator outside: err = %v, want ErrForbidden", err)
	}
	if _, err := operation.Do(ctx, f.r, AddMember, f.olivier, MemberInput{Room: "marketing", User: "olivier", Role: Owner}); err != nil {
		t.Errorf("add_room_member by an administrator: %v", err)
	}
	if got := f.people(t, f.camille, "marketing"); !slices.Equal(got, []string{"camille owner", "olivier owner", "bruno member"}) {
		t.Errorf("people of marketing = %q", got)
	}

	// A Room keeps an owner.
	if _, err := operation.Do(ctx, f.r, SetMemberRole, f.camille, MemberInput{Room: "marketing", User: "olivier", Role: Member}); err != nil {
		t.Fatalf("set_room_member_role: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, SetMemberRole, f.camille, MemberInput{Room: "marketing", User: "camille", Role: Member}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("demote the last owner: err = %v, want ErrInvalid", err)
	}
	if _, err := operation.Do(ctx, f.r, RemoveMember, f.camille, RemoveMemberInput{Room: "marketing", User: "camille"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("the last owner leaves: err = %v, want ErrInvalid", err)
	}
	if _, err := operation.Do(ctx, f.r, SetMemberRole, f.camille, MemberInput{Room: "marketing", User: "nobody", Role: Member}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("set_room_member_role of nobody: err = %v, want ErrNotFound", err)
	}

	// Anyone leaves a Room; only owners remove others.
	if _, err := operation.Do(ctx, f.r, RemoveMember, f.bruno, RemoveMemberInput{Room: "marketing", User: "olivier"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("a member removes another: err = %v, want ErrForbidden", err)
	}
	if _, err := operation.Do(ctx, f.r, RemoveMember, f.bruno, RemoveMemberInput{Room: "marketing", User: "bruno"}); err != nil {
		t.Errorf("a member leaves: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, RemoveMember, f.camille, RemoveMemberInput{Room: "marketing", User: "olivier"}); err != nil {
		t.Errorf("an owner removes a member: %v", err)
	}
	if got := f.people(t, f.camille, "marketing"); !slices.Equal(got, []string{"camille owner"}) {
		t.Errorf("people of marketing = %q", got)
	}
	if _, err := operation.Do(ctx, f.r, RemoveMember, f.camille, RemoveMemberInput{Room: "camille", User: "camille"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("leave a personal Room: err = %v, want ErrInvalid", err)
	}

	// Every change is in the Logbook, under the Room.
	entries, err := logbook.List(ctx, f.db.Read, 100)
	if err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := f.db.Read.QueryRowContext(ctx, "SELECT id FROM rooms WHERE name = 'marketing'").Scan(&id); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range entries {
		if e.RoomID == id && e.Outcome == logbook.OK {
			n++
		}
	}
	// The creation, two additions, a role, a departure and a removal; the
	// reads of people are not logged.
	if n != 6 {
		t.Errorf("%d entries of marketing, want 6", n)
	}
}

// call returns a call of the person p, in a read-only transaction, to test
// the helpers that operations use.
func call(t *testing.T, f *fixture, p operation.Actor) *operation.Call {
	t.Helper()
	tx, err := f.db.Read.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	admin, err := account.Check(context.Background(), tx, p.UserID)
	if err != nil {
		t.Fatal(err)
	}
	return &operation.Call{Actor: p, Tx: tx, Admin: admin}
}
