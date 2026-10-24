package cahut

import (
	"context"
	"errors"
	"slices"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/window"
)

func TestListLogbook(t *testing.T) {
	ctx := context.Background()
	c, err := Open(ctx, Config{Dir: t.TempDir(), Domain: "cahut.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	admin := operation.Actor{Kind: operation.Admin, Via: "cli"}
	people := map[string]operation.Actor{}
	for _, handle := range []string{"olivier", "camille", "bruno"} {
		u, err := operation.Do(ctx, c.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: handle, Name: handle, Admin: handle == "olivier"})
		if err != nil {
			t.Fatal(err)
		}
		people[handle] = operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: handle, Session: 1}
		if _, err := operation.Do(ctx, c.Registry, account.CreateToken, people[handle], account.CreateTokenInput{Name: "Claude Code"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := operation.Do(ctx, c.Registry, room.CreateRoom, people["camille"], room.CreateRoomInput{Title: "Marketing"}); err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, c.Registry, room.AddMember, people["camille"], room.MemberInput{Room: "marketing", User: "bruno"}); err != nil {
		t.Fatal(err)
	}

	targets := func(who operation.Actor) []string {
		t.Helper()
		page, err := operation.Do(ctx, c.Registry, ListLogbook, who, ListLogbookInput{})
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, e := range page.Entries {
			got = append(got, e.Operation+" "+e.Target)
		}
		return got
	}
	// An administrator reads everything; anyone else, what they did, what
	// their AIs did, and what happened in their Rooms.
	if got := targets(people["olivier"]); len(got) != 8 {
		t.Errorf("the administrator reads %q, want 8 entries", got)
	}
	want := []string{"add_room_member marketing", "create_room marketing", "create_token token 3 of bruno", "create_user bruno"}
	if got := targets(people["bruno"]); !slices.Equal(got, want) {
		t.Errorf("bruno reads %q, want %q", got, want)
	}
	if _, err := operation.Do(ctx, c.Registry, ListLogbook, admin, ListLogbookInput{Limit: 1001}); err == nil {
		t.Error("list_logbook with a limit above 1000 succeeded")
	}
}

func TestGetCahut(t *testing.T) {
	ctx := context.Background()
	c, err := Open(ctx, Config{Dir: t.TempDir(), Domain: "cahut.example.com", BaseURL: "https://cahut.example.com/"})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	admin := operation.Actor{Kind: operation.Admin, Via: "cli"}
	u, err := operation.Do(ctx, c.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: "olivier", Name: "Olivier"})
	if err != nil {
		t.Fatal(err)
	}
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: u.ID, Handle: "olivier", TokenID: 1, Client: "Claude"}
	out, err := c.Registry.Invoke(ctx, "get_cahut", ai, nil)
	want := About{Domain: "cahut.example.com", URL: "https://cahut.example.com/", MCP: "https://cahut.example.com/mcp",
		Settings: settings.Settings{Usage: settings.Usage{People: 1}}}
	if err != nil || out != want {
		t.Errorf("get_cahut = %+v, %v; want %+v", out, err, want)
	}
}

// TestReach checks that an AI reaches only what its person granted it: some
// Rooms, and reading only. See docs/adr/0016-oauth-for-ai-clients.md.
func TestReach(t *testing.T) {
	ctx := context.Background()
	c, err := Open(ctx, Config{Dir: t.TempDir(), Domain: "cahut.example.com", BaseURL: "https://cahut.example.com/"})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	admin := operation.Actor{Kind: operation.Admin, Via: "cli"}
	u, err := operation.Do(ctx, c.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: "olivier", Name: "Olivier", Admin: true})
	if err != nil {
		t.Fatal(err)
	}
	person := operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: "olivier", Session: 1}
	if _, err := operation.Do(ctx, c.Registry, room.CreateRoom, person, room.CreateRoomInput{Title: "Marketing"}); err != nil {
		t.Fatal(err)
	}
	var marketing int64
	if err := c.DB.Read.QueryRowContext(ctx, "SELECT id FROM rooms WHERE name = 'marketing'").Scan(&marketing); err != nil {
		t.Fatal(err)
	}
	full := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: u.ID, Handle: "olivier", TokenID: 1, Client: "Claude"}
	for _, in := range []string{
		`{"title": "Plan", "kind": "document", "room": "olivier", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<h1>Plan</h1>"}]}`,
		`{"title": "Q3", "kind": "document", "room": "marketing", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<h1>Q3</h1>"}]}`,
	} {
		if _, err := c.Registry.Invoke(ctx, "publish_window", full, []byte(in)); err != nil {
			t.Fatal(err)
		}
	}

	limited := full
	limited.TokenID, limited.Rooms = 2, []int64{marketing}
	call := func(ai operation.Actor, name, in string) (any, error) {
		t.Helper()
		return c.Registry.Invoke(ctx, name, ai, []byte(in))
	}
	out, err := call(limited, "list_windows", `{}`)
	if err != nil || len(out.(window.List).Windows) != 1 || out.(window.List).Windows[0].Room != "marketing" {
		t.Errorf("list_windows by an AI limited to marketing = %+v, %v", out, err)
	}
	out, err = call(limited, "list_rooms", `{"all": true}`)
	if err != nil || len(out.(room.List).Rooms) != 1 {
		t.Errorf("list_rooms by an AI limited to marketing = %+v, %v", out, err)
	}
	out, err = call(limited, "get_account", `{}`)
	if err != nil || len(out.(account.Account).Rooms) != 1 {
		t.Errorf("get_account by an AI limited to marketing = %+v, %v", out, err)
	}
	for name, in := range map[string]string{
		"get_window":     `{"window": "olivier/plan"}`,
		"get_room":       `{"room": "olivier"}`,
		"list_windows":   `{"room": "olivier"}`,
		"publish_window": `{"title": "Other", "kind": "document", "room": "olivier", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "x"}]}`,
	} {
		if _, err := call(limited, name, in); !errors.Is(err, operation.ErrNotFound) {
			t.Errorf("%s outside the granted Rooms: %v, want ErrNotFound", name, err)
		}
	}
	if _, err := call(limited, "create_room", `{"title": "Sales"}`); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("create_room by an AI limited to some Rooms: %v, want ErrForbidden", err)
	}
	page, err := operation.Do(ctx, c.Registry, ListLogbook, limited, ListLogbookInput{})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range page.Entries {
		if e.RoomID != marketing {
			t.Errorf("an AI limited to marketing reads %s %s, about Room %d", e.Operation, e.Target, e.RoomID)
		}
	}
	if len(page.Entries) == 0 {
		t.Error("an AI limited to marketing reads nothing of it in the Logbook")
	}

	readOnly := full
	readOnly.TokenID, readOnly.ReadOnly = 3, true
	if _, err := call(readOnly, "publish_window", `{"title": "Q4", "kind": "document", "room": "marketing", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "x"}]}`); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("publish_window by an AI that may only read: %v, want ErrForbidden", err)
	}
	out, err = call(readOnly, "list_windows", `{}`)
	if err != nil || len(out.(window.List).Windows) != 2 {
		t.Errorf("list_windows by an AI that may only read = %+v, %v", out, err)
	}
}
