package knock_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/knock"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/oauth"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/window"
)

var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

type fixture struct {
	c   *cahut.Cahut
	now time.Time
	// olivier is an administrator of the Cahut; camille is a member.
	olivier, camille operation.Actor // in the console
	ai               operation.Actor // the AI of olivier, with a personal token
}

func setup(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	c, err := cahut.Open(ctx, cahut.Config{Dir: t.TempDir(), Domain: "cahut.test",
		BaseURL: "https://cahut.test/", ConsoleURL: "https://cahut.test/console/"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	f := &fixture{c: c, now: time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)}
	c.Registry.Now = func() time.Time { return f.now }
	ids := map[string]int64{}
	for _, in := range []account.CreateUserInput{{Handle: "olivier", Name: "Olivier Durand", Admin: true}, {Handle: "camille", Name: "Camille"}} {
		u, err := operation.Do(ctx, c.Registry, account.CreateUser, admin, in)
		if err != nil {
			t.Fatal(err)
		}
		ids[in.Handle] = u.ID
	}
	tok, err := operation.Do(ctx, c.Registry, account.CreateToken, admin, account.CreateTokenInput{User: "olivier", Name: "Claude Code"})
	if err != nil {
		t.Fatal(err)
	}
	f.olivier = operation.Actor{Kind: operation.Person, Via: "api", UserID: ids["olivier"], Handle: "olivier", Session: 1}
	f.camille = operation.Actor{Kind: operation.Person, Via: "api", UserID: ids["camille"], Handle: "camille", Session: 2}
	f.ai = operation.Actor{Kind: operation.AI, Via: "mcp", UserID: ids["olivier"], Handle: "olivier", TokenID: tok.ID, Client: "Claude Code"}
	return f
}

func (f *fixture) invoke(t *testing.T, actor operation.Actor, op, args string) (any, error) {
	t.Helper()
	return f.c.Registry.Invoke(context.Background(), op, actor, json.RawMessage(args))
}

// publish publishes a Window in the personal Room of olivier, and returns
// its address.
func (f *fixture) publish(t *testing.T, title string) string {
	t.Helper()
	out, err := f.invoke(t, f.olivier, "publish_window", `{"title": "`+title+`", "kind": "deck",
		"files": [{"path": "index.html", "content": "<h1>Q3</h1>"}], "declaration": {"scripts": false}}`)
	if err != nil {
		t.Fatal(err)
	}
	return out.(window.Window).URL
}

func (f *fixture) access(t *testing.T, url string) string {
	t.Helper()
	out, err := f.invoke(t, f.olivier, "get_window", `{"window": "`+url+`"}`)
	if err != nil {
		t.Fatal(err)
	}
	return out.(window.Detail).Access
}

// latest returns the latest entry of the Logbook for the operation op.
func (f *fixture) latest(t *testing.T, op string) logbook.Entry {
	t.Helper()
	entries, err := logbook.List(context.Background(), f.c.DB.Read, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Operation == op {
			return e
		}
	}
	t.Fatalf("no %s in the Logbook", op)
	return logbook.Entry{}
}

// knocked returns the Knock that err holds.
func knocked(t *testing.T, err error) *operation.Knocked {
	t.Helper()
	var k *operation.Knocked
	if !errors.As(err, &k) {
		t.Fatalf("err = %v, want a Knock", err)
	}
	return k
}

func (f *fixture) get(t *testing.T, actor operation.Actor, id int64) knock.Knock {
	t.Helper()
	out, err := f.invoke(t, actor, "get_knock", fmt.Sprintf(`{"knock": %d}`, id))
	if err != nil {
		t.Fatalf("get_knock %d: %v", id, err)
	}
	return out.(knock.Knock)
}

func (f *fixture) answer(t *testing.T, actor operation.Actor, op string, id int64) (any, error) {
	t.Helper()
	return f.invoke(t, actor, op, fmt.Sprintf(`{"knock": %d}`, id))
}

func TestApprove(t *testing.T) {
	f := setup(t)
	url := f.publish(t, "Q3 review")

	_, err := f.invoke(t, f.ai, "set_window_access", `{"window": "`+url+`", "access": "public"}`)
	k := knocked(t, err)
	if k.URL != fmt.Sprintf("https://cahut.test/console/knocks/%d", k.ID) || k.Person != "Olivier Durand" ||
		!k.ExpiresAt.Equal(f.now.Add(knock.Lifetime)) {
		t.Errorf("Knock = %+v", k)
	}
	if got := f.access(t, url); got != window.AccessRoom {
		t.Fatalf("the Knock made the Window %s", got)
	}

	// The AI and its person read the Knock, which waits.
	got := f.get(t, f.ai, k.ID)
	if got.State != knock.Waiting || got.Operation != "set_window_access" || got.Input["access"] != "public" ||
		got.Target != "olivier/q3-review" || got.Room != "olivier" || got.User != "olivier" || got.Client != "Claude Code" ||
		got.URL != k.URL || got.AnsweredAt != nil {
		t.Errorf("get_knock = %+v", got)
	}
	out, err := f.invoke(t, f.olivier, "list_knocks", `{}`)
	if l := out.(knock.List); err != nil || l.Waiting != 1 || len(l.Knocks) != 1 || l.Knocks[0].ID != k.ID {
		t.Errorf("list_knocks = %+v, %v", out, err)
	}

	f.now = f.now.Add(time.Hour)
	out, err = f.answer(t, f.olivier, "approve_knock", k.ID)
	if err != nil {
		t.Fatalf("approve_knock: %v", err)
	}
	a := out.(knock.Answer)
	if a.Knock.State != knock.Done || a.Knock.AnsweredAt == nil || !a.Knock.AnsweredAt.Equal(f.now) ||
		a.Output.(window.AccessSet).Access != window.AccessPublic {
		t.Errorf("approve_knock = %+v", a)
	}
	if got := f.access(t, url); got != window.AccessPublic {
		t.Errorf("approved, the Window opens to %s", got)
	}
	if got := f.get(t, f.ai, k.ID); got.State != knock.Done {
		t.Errorf("the AI reads %s", got.State)
	}
	if _, err := f.answer(t, f.olivier, "approve_knock", k.ID); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("a second approval: err = %v, want ErrInvalid", err)
	}

	// The Logbook shows the Knock, then the call, as the AI's, through the
	// console.
	entries, err := logbook.List(context.Background(), f.c.DB.Read, 10)
	if err != nil {
		t.Fatal(err)
	}
	var steps []string
	for _, e := range entries {
		if e.KnockID == k.ID {
			steps = append(steps, fmt.Sprintf("%s %s %s %s", e.Operation, e.Actor, e.Via, e.Outcome))
		}
	}
	want := []string{"approve_knock person api failed", "get_knock ai mcp ok", "set_window_access ai api ok",
		"get_knock ai mcp ok", "set_window_access ai mcp knocked"}
	if strings.Join(steps, "; ") != strings.Join(want, "; ") {
		t.Errorf("Logbook:\n%s\nwant\n%s", strings.Join(steps, "\n"), strings.Join(want, "\n"))
	}
}

func TestDeny(t *testing.T) {
	f := setup(t)
	url := f.publish(t, "Q3 review")
	_, err := f.invoke(t, f.ai, "set_window_access", `{"window": "`+url+`", "access": "public"}`)
	k := knocked(t, err)

	out, err := f.answer(t, f.olivier, "deny_knock", k.ID)
	if err != nil || out.(knock.Knock).State != knock.Denied {
		t.Fatalf("deny_knock = %+v, %v", out, err)
	}
	if got := f.get(t, f.ai, k.ID); got.State != knock.Denied || f.access(t, url) != window.AccessRoom {
		t.Errorf("denied: %s, the Window opens to %s", got.State, f.access(t, url))
	}
	if _, err := f.answer(t, f.olivier, "approve_knock", k.ID); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("approving a denied Knock: err = %v", err)
	}
	if e := f.latest(t, "deny_knock"); e.KnockID != k.ID || e.Actor != "person" || e.Outcome != logbook.OK {
		t.Errorf("Logbook: %+v", e)
	}
}

func TestOnlyThePersonAnswers(t *testing.T) {
	f := setup(t)
	url := f.publish(t, "Q3 review")
	_, err := f.invoke(t, f.ai, "set_window_access", `{"window": "`+url+`", "access": "public"}`)
	k := knocked(t, err)

	for _, op := range []string{"approve_knock", "deny_knock"} {
		if _, err := f.answer(t, f.ai, op, k.ID); !errors.Is(err, operation.ErrForbidden) {
			t.Errorf("%s by the AI: err = %v, want ErrForbidden", op, err)
		}
		if _, err := f.answer(t, f.camille, op, k.ID); !errors.Is(err, operation.ErrNotFound) {
			t.Errorf("%s by another person: err = %v, want ErrNotFound", op, err)
		}
		if _, err := f.answer(t, admin, op, k.ID); !errors.Is(err, operation.ErrForbidden) {
			t.Errorf("%s by the command line: err = %v, want ErrForbidden", op, err)
		}
	}
	if _, err := f.invoke(t, f.camille, "get_knock", fmt.Sprintf(`{"knock": %d}`, k.ID)); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_knock by another person: err = %v", err)
	}
	out, err := f.invoke(t, f.camille, "list_knocks", `{}`)
	if l := out.(knock.List); err != nil || l.Waiting != 0 || len(l.Knocks) != 0 {
		t.Errorf("list_knocks of another person = %+v, %v", out, err)
	}
	if f.get(t, f.olivier, k.ID).State != knock.Waiting {
		t.Error("the Knock was answered")
	}
}

func TestExpiry(t *testing.T) {
	f := setup(t)
	url := f.publish(t, "Q3 review")
	_, err := f.invoke(t, f.ai, "set_window_access", `{"window": "`+url+`", "access": "public"}`)
	k := knocked(t, err)

	f.now = f.now.Add(knock.Lifetime)
	if got := f.get(t, f.ai, k.ID); got.State != knock.Expired {
		t.Errorf("after its lifetime, the Knock is %s", got.State)
	}
	if _, err := f.answer(t, f.olivier, "approve_knock", k.ID); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("approving an expired Knock: err = %v", err)
	}
	// The same call makes a new Knock.
	_, err = f.invoke(t, f.ai, "set_window_access", `{"window": "`+url+`", "access": "public"}`)
	if again := knocked(t, err); again.ID == k.ID {
		t.Error("an expired Knock was returned again")
	}
}

func TestCanceledWhenTheAILosesAccess(t *testing.T) {
	f := setup(t)
	url := f.publish(t, "Q3 review")
	_, err := f.invoke(t, f.ai, "set_window_access", `{"window": "`+url+`", "access": "public"}`)
	k := knocked(t, err)

	if _, err := f.invoke(t, f.olivier, "revoke_token", fmt.Sprintf(`{"id": %d}`, f.ai.TokenID)); err != nil {
		t.Fatal(err)
	}
	if got := f.get(t, f.olivier, k.ID); got.State != knock.Canceled {
		t.Errorf("after the revocation of its token, the Knock is %s", got.State)
	}
	if _, err := f.answer(t, f.olivier, "approve_knock", k.ID); !errors.Is(err, operation.ErrInvalid) || f.access(t, url) != window.AccessRoom {
		t.Errorf("approving a canceled Knock: err = %v", err)
	}
}

func TestSameCallSameKnock(t *testing.T) {
	f := setup(t)
	url := f.publish(t, "Q3 review")
	call := func(access string) *operation.Knocked {
		t.Helper()
		_, err := f.invoke(t, f.ai, "set_window_access", `{"window": "`+url+`", "access": "`+access+`"}`)
		return knocked(t, err)
	}
	first := call("public")
	if again := call("public"); again.ID != first.ID {
		t.Errorf("the same call made Knock %d, then %d", first.ID, again.ID)
	}
	if other, err := f.invoke(t, f.ai, "disable_user", `{"user": "camille"}`); err == nil || knocked(t, err).ID == first.ID {
		t.Errorf("another call: %v, %v", other, err)
	}
	out, err := f.invoke(t, f.olivier, "list_knocks", `{"waiting": true}`)
	if l := out.(knock.List); err != nil || l.Waiting != 2 || len(l.Knocks) != 2 {
		t.Errorf("list_knocks = %+v, %v", out, err)
	}
}

func TestMaxWaiting(t *testing.T) {
	f := setup(t)
	for i := range knock.MaxWaiting {
		_, err := f.invoke(t, f.ai, "create_user", fmt.Sprintf(`{"handle": "person-%d", "name": "Person %d"}`, i, i))
		_ = knocked(t, err)
	}
	_, err := f.invoke(t, f.ai, "create_user", `{"handle": "one-more", "name": "One more"}`)
	if !errors.Is(err, operation.ErrTooMany) {
		t.Errorf("err = %v, want ErrTooMany", err)
	}
	entries, err := logbook.List(context.Background(), f.c.DB.Read, 1)
	if err != nil || entries[0].Outcome != logbook.Refused {
		t.Errorf("Logbook = %+v, %v", entries, err)
	}
}

func TestApprovedCreateUserGivesTheLinkToThePerson(t *testing.T) {
	f := setup(t)
	_, err := f.invoke(t, f.ai, "create_user", `{"handle": "bruno", "name": "Bruno"}`)
	k := knocked(t, err)

	out, err := f.answer(t, f.olivier, "approve_knock", k.ID)
	if err != nil {
		t.Fatal(err)
	}
	u := out.(knock.Answer).Output.(account.NewUser)
	if u.Handle != "bruno" || u.Activation == nil || u.Activation.Code == "" {
		t.Errorf("approved, create_user gives %+v", u)
	}
	got, err := json.Marshal(f.get(t, f.ai, k.ID))
	if err != nil || strings.Contains(string(got), u.Activation.Code) {
		t.Errorf("the AI reads the activation code: %s", got)
	}
}

func TestApprovalThatFails(t *testing.T) {
	f := setup(t)
	_, err := f.invoke(t, f.ai, "create_user", `{"handle": "bruno", "name": "Bruno"}`)
	k := knocked(t, err)
	// Meanwhile, someone takes the handle.
	if _, err := f.invoke(t, f.olivier, "create_user", `{"handle": "bruno", "name": "Bruno B."}`); err != nil {
		t.Fatal(err)
	}

	out, err := f.answer(t, f.olivier, "approve_knock", k.ID)
	if err != nil {
		t.Fatalf("approve_knock: %v", err)
	}
	a := out.(knock.Answer)
	if a.Knock.State != knock.Failed || !strings.Contains(a.Knock.Error, `handle "bruno" is taken`) || a.Output != nil {
		t.Errorf("approve_knock = %+v", a)
	}
	if got := f.get(t, f.ai, k.ID); got.State != knock.Failed || got.Error != a.Knock.Error {
		t.Errorf("get_knock = %+v", got)
	}
	if e := f.latest(t, "create_user"); e.Outcome != logbook.Failed || e.KnockID != k.ID || e.Actor != "ai" {
		t.Errorf("Logbook: %+v", e)
	}
}

// TestApprovalChecksTheReachOfTheAI connects an AI to one Room, has it
// knock, then takes the Room away: the approval runs with the reach the AI
// has then.
func TestApprovalChecksTheReachOfTheAI(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	url := f.publish(t, "Q3 review")
	db := f.c.DB.Write
	conn, ai := f.connect(t, "write", "olivier")

	_, err := f.invoke(t, ai, "set_window_access", `{"window": "`+url+`", "access": "public"}`)
	k := knocked(t, err)
	if _, err := db.ExecContext(ctx, "DELETE FROM connection_rooms WHERE connection_id = ?", conn); err != nil {
		t.Fatal(err)
	}

	out, err := f.answer(t, f.olivier, "approve_knock", k.ID)
	if err != nil {
		t.Fatal(err)
	}
	if a := out.(knock.Answer); a.Knock.State != knock.Failed || f.access(t, url) != window.AccessRoom {
		t.Errorf("approve_knock = %+v", a)
	}

	// A connection that ends cancels its Knocks.
	_, err = f.invoke(t, ai, "disable_user", `{"user": "camille"}`)
	k = knocked(t, err)
	if _, err := db.ExecContext(ctx, "UPDATE connections SET ended_at = ? WHERE id = ?", store.FormatTime(f.now), conn); err != nil {
		t.Fatal(err)
	}
	if got := f.get(t, f.olivier, k.ID); got.State != knock.Canceled {
		t.Errorf("after its connection ended, the Knock is %s", got.State)
	}
}

// connect connects an AI for olivier, as the consent screen would, with
// the scope and the Rooms given. It returns the connection and the AI.
func (f *fixture) connect(t *testing.T, scope string, rooms ...string) (int64, operation.Actor) {
	t.Helper()
	ctx := context.Background()
	db := f.c.DB.Write
	var conn int64
	err := db.QueryRowContext(ctx, `
		INSERT INTO connections (user_id, client_id, client_name, redirect_host, scope, all_rooms, created_at)
		VALUES (?, 'https://claude.ai/client', 'Claude', 'claude.ai', ?, 0, ?) RETURNING id`,
		f.olivier.UserID, scope, store.FormatTime(f.now)).Scan(&conn)
	if err != nil {
		t.Fatal(err)
	}
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: f.olivier.UserID, Handle: "olivier", Conn: conn,
		Client: "Claude", ReadOnly: scope == "read", Rooms: []int64{}}
	for _, name := range rooms {
		var id int64
		if err := db.QueryRowContext(ctx, "SELECT id FROM rooms WHERE name = ?", name).Scan(&id); err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, "INSERT INTO connection_rooms (connection_id, room_id) VALUES (?, ?)", conn, id); err != nil {
			t.Fatal(err)
		}
		ai.Rooms = append(ai.Rooms, id)
	}
	return conn, ai
}

func (f *fixture) reach(t *testing.T, conn int64) oauth.Connection {
	t.Helper()
	out, err := f.invoke(t, f.olivier, "list_connected_ais", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range out.(oauth.ConnectedAIs).Connections {
		if c.ID == conn {
			return c
		}
	}
	t.Fatalf("no connection %d", conn)
	return oauth.Connection{}
}

func TestSetAIReach(t *testing.T) {
	f := setup(t)
	if _, err := f.invoke(t, f.olivier, "create_room", `{"title": "Marketing"}`); err != nil {
		t.Fatal(err)
	}
	conn, ai := f.connect(t, "read", "olivier")

	// An AI that may only read asks to write: it knocks.
	_, err := f.invoke(t, ai, "set_ai_reach", `{"scope": "write", "rooms": ["olivier"]}`)
	write := knocked(t, err)
	if r := f.reach(t, conn); r.Scope != "read" {
		t.Errorf("the Knock gave %s", r.Scope)
	}
	// It names a Room it does not reach, and one that does not exist: it
	// learns nothing of them before its person answers.
	_, err = f.invoke(t, ai, "set_ai_reach", `{"scope": "read", "rooms": ["olivier", "nope"]}`)
	nope := knocked(t, err)

	out, err := f.answer(t, f.olivier, "approve_knock", write.ID)
	if err != nil || out.(knock.Answer).Knock.State != knock.Done {
		t.Fatalf("approve_knock = %+v, %v", out, err)
	}
	if r := f.reach(t, conn); r.Scope != "write" || r.AllRooms || len(r.Rooms) != 1 || r.Rooms[0] != "olivier" {
		t.Errorf("approved, the AI reaches %+v", r)
	}
	out, err = f.answer(t, f.olivier, "approve_knock", nope.ID)
	if a := out.(knock.Answer); err != nil || a.Knock.State != knock.Failed || !strings.Contains(a.Knock.Error, `no Room "nope"`) {
		t.Errorf("approving a Room that does not exist = %+v, %v", out, err)
	}

	// The person sets it at once; the AI narrows it at once.
	out, err = f.invoke(t, f.olivier, "set_ai_reach", fmt.Sprintf(`{"id": %d, "scope": "write", "rooms": ["olivier", "marketing"]}`, conn))
	if r := out.(oauth.Connection); err != nil || len(r.Rooms) != 2 {
		t.Fatalf("set_ai_reach by the person = %+v, %v", out, err)
	}
	ai.ReadOnly = false
	out, err = f.invoke(t, ai, "set_ai_reach", `{"scope": "read", "rooms": ["marketing"]}`)
	if r := out.(oauth.Connection); err != nil || r.Scope != "read" || len(r.Rooms) != 1 || r.Rooms[0] != "marketing" {
		t.Errorf("set_ai_reach that narrows = %+v, %v", out, err)
	}
	if _, err := f.invoke(t, ai, "set_ai_reach", `{"scope": "read", "all_rooms": true}`); !errors.Is(err, operation.ErrApprovalRequired) {
		t.Errorf("every Room: err = %v, want a Knock", err)
	}

	for _, tt := range []struct {
		name  string
		actor operation.Actor
		args  string
		want  error
	}{
		{name: "a personal token", actor: f.ai, args: `{"scope": "read", "all_rooms": true}`, want: operation.ErrInvalid},
		{name: "another connection", actor: ai, args: fmt.Sprintf(`{"id": %d, "scope": "read", "all_rooms": true}`, conn+1), want: operation.ErrForbidden},
		{name: "another person", actor: f.camille, args: fmt.Sprintf(`{"id": %d, "scope": "read", "all_rooms": true}`, conn), want: operation.ErrNotFound},
		{name: "no Room", actor: f.olivier, args: fmt.Sprintf(`{"id": %d, "scope": "read"}`, conn), want: operation.ErrInvalid},
		{name: "a bad scope", actor: f.olivier, args: fmt.Sprintf(`{"id": %d, "scope": "admin", "all_rooms": true}`, conn), want: operation.ErrInvalid},
	} {
		if _, err := f.invoke(t, tt.actor, "set_ai_reach", tt.args); !errors.Is(err, tt.want) {
			t.Errorf("%s: err = %v, want %v", tt.name, err, tt.want)
		}
	}
}
